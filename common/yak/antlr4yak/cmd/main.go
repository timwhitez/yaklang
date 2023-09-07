package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func main() {
	code, err := os.ReadFile("./marshalCode")
	if err != nil {
		spew.Dump(err)
		panic("read code error")
	}
	m := yakvm.NewCodesMarshaller()
	symbolTable, codes, err := m.Unmarshal(code)
	if err != nil {
		panic("unmarshal: " + err.Error())
	}
	vm := yakvm.NewWithSymbolTable(symbolTable)
	vm.SetVar("print", func(args ...interface{}) {
		fmt.Print(args...)
	})
	vm.SetVar("println", func(args ...interface{}) {
		fmt.Println(args...)
	})
	vm.SetVar("get", func() []byte {
		var input string
		fmt.Scanln(&input)
		return []byte(input)
	})
	vm.SetVar("string2int", func(str string) int {
		i, err := strconv.Atoi(str)
		if err != nil {
			return 0
		}
		return i
	})
	vm.SetVar("len", func(a []byte) int {
		return len(a)
	})
	err = vm.Exec(context.Background(), func(frame *yakvm.Frame) {
		frame.NormalExec(codes)
	})
	if err != nil {
		spew.Dump(err)
		panic("exec code error")
	}
}
