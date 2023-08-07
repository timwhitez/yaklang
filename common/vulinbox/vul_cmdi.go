package vulinbox

import (
	"context"
	"fmt"
	"github.com/google/shlex"
	"github.com/yaklang/yaklang/common/utils"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func (s *VulinServer) registerPingCMDI() {
	r := s.router

	cmdIGroup := r.PathPrefix("/exec").Name("命令注入测试案例 (Unsafe Mode)").Subrouter()
	cmdIRoutes := []*VulInfo{
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/interactive/shlex",
			Title:        "Shlex  解析的命令注入(完全参数化)",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte("ip is empty"))
					return
				}
				outputs, err := shlexDo(ip)
				if err != nil {
					writer.Write([]byte(err.Error()))
				} else {
					writer.Write(outputs)
				}
			},
			RiskDetected: false,
		},
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/interactive/nowaf/cmdline",
			Title:        "命令行解析的命令注入(无waf)",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte("ip is empty"))
					return
				}
				outputs, err := cmdlineDo(ip)
				if err != nil {
					writer.Write([]byte(err.Error()))
				} else {
					writer.Write(outputs)
				}

			},
			RiskDetected: true,
		},
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/interactive/nowaf/cmdline",
			Title:        "命令行解析的命令注入(无waf)",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte("ip is empty"))
					return
				}
				outputs, err := cmdlineDo(ip)
				if err != nil {
					writer.Write([]byte(err.Error()))
				} else {
					writer.Write(outputs)
				}

			},
			RiskDetected: true,
		},
		{
			DefaultQuery: "ip=127.0.0.1",
			Path:         "/interactive/waf/level1/cmdline",
			Title:        "命令行解析的命令注入(过滤空格)",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				ip := request.URL.Query().Get("ip")
				if ip == "" {
					writer.Write([]byte("ip is empty"))
					return
				}
				ip = filterBlank(ip)
				outputs, err := cmdlineDo(ip)
				if err != nil {
					writer.Write([]byte(err.Error()))
				} else {
					writer.Write(outputs)
				}

			},
			RiskDetected: true,
		},
	}

	for _, v := range cmdIRoutes {
		addRouteWithVulInfo(cmdIGroup, v)
	}
}

func filterBlank(cmd string) string {
	return strings.Replace(cmd, " ", "", -1)
}

func filterSymbol(cmd string) string {
	replacer := strings.NewReplacer(";", "", "|", "", "||", "", "&", "", "&&", "", "`", "")
	return replacer.Replace(cmd)
}

func filterSymbolIncomplete(cmd string) string {
	replacer := strings.NewReplacer(";", "", "|", "", "&", "")
	return replacer.Replace(cmd)
}

func cmdlineDo(ip string) ([]byte, error) {
	var cmdline string
	var arg string
	switch runtime.GOOS {
	case "linux":
		cmdline = "bash"
		cmd := exec.Command("bash", "-c", "echo Bash is installed!")
		err := cmd.Run()
		if err != nil {
			cmdline = "sh"
		}
		arg = "-c"
	case "windows":
		cmdline = "cmd"
		arg = "/C"
	}
	var raw = fmt.Sprintf("ping %v", ip)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	outputs, err := exec.CommandContext(ctx, cmdline, arg, raw).CombinedOutput()
	if err != nil {
		return nil, err
	}
	// 尝试将 GBK 转换为 UTF-8
	utf8Outputs, err := utils.GbkToUtf8(outputs)
	if err != nil {
		return outputs, nil
	}
	return utf8Outputs, nil
}

func shlexDo(ip string) ([]byte, error) {
	var raw = fmt.Sprintf("ping %v", ip)
	list, err := shlex.Split(raw)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	outputs, err := exec.CommandContext(ctx, list[0], list[1:]...).CombinedOutput()
	if err != nil {
		return nil, err
	}
	// 尝试将 GBK 转换为 UTF-8
	utf8Outputs, err := utils.GbkToUtf8(outputs)
	if err != nil {
		return outputs, nil
	}
	return utf8Outputs, nil
}
