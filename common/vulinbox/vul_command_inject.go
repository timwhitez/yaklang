package vulinbox

import (
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func (s *VulinServer) registerCommandInject() {
	router := s.router

	router.HandleFunc("/cmdi/system/join", func(writer http.ResponseWriter, request *http.Request) {
		//var c = request.URL.Query().Get("ping")

	})
}

func filter(command string, level int) {

}

func commandExec(command string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		err := exec.Command("chcp", "65001").Run()
		if err != nil {
			return nil, err
		}
	}
	cmdArray := strings.FieldsFunc(command, func(r rune) bool {
		return r == '\n' || r == '\f' || r == '\t' || r == '\r' || r == ' '
	})

	cmd := exec.Command(cmdArray[0], cmdArray[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func systemExec(command string) ([]byte, error) {
	var commandLine string
	var arg string
	switch runtime.GOOS {
	case "linux":
		commandLine = "bash"
		cmd := exec.Command("bash", "-c", "echo Bash is installed!")
		err := cmd.Run()
		if err != nil {
			commandLine = "sh"
		}
		arg = "-c"
	case "windows":
		commandLine = "cmd"
		arg = "/C"
		err := exec.Command("chcp", "65001").Run()
		if err != nil {
			return nil, err
		}
	}
	cmd := exec.Command(commandLine, arg, command)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}
