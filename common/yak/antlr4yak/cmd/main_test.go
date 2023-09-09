package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakast"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

func Encode(flag []byte) {
	for i, v := range flag {
		if i%2 == 0 {
			v = v ^ 0xf0
		} else {
			v = v ^ 0x0f
		}
		v = v ^ byte(2*i)
		flag[i] = v
	}
}

func Decode(flag []byte) {
	for i, v := range flag {
		v = v ^ byte(2*i)
		if i%2 == 0 {
			v = v ^ 0xf0
		} else {
			v = v ^ 0x0f
		}
		flag[i] = v
	}
}

func TestReverse(t *testing.T) {
	bytes, err := os.ReadFile("./marshalCode")
	if err != nil {
		panic("read file error")
	}

	m := yakvm.NewCodesMarshaller()
	table, code, err := m.Unmarshal(bytes)
	if err != nil {
		panic("unmarshal error")
	}

	yakvm.ShowOpcodes(code)

	// run
	vm := yakvm.NewWithSymbolTable(table)
	vm.SetVar("print", func(arg ...interface{}) {
		fmt.Println(arg)
	})
	vm.SetVar("println", func() {
		fmt.Println("println")
	})

	vm.SetVar("get", func() {
		fmt.Println("get")
	})

	vm.SetVar("string2int", func() {
		fmt.Println("string2int")
	})

	vm.SetVar("len", func() {
		fmt.Println("len")
	})

	// vm.Exec(context.Background(), func(frame *yakvm.Frame) {
	// 	frame.NormalExec(code)
	// })
}

func TestCheckFlag(t *testing.T) {
	flag := []byte("flag{aaaaaaaaaaaaaaaaa}")
	fmt.Println(flag)
	Encode(flag)
	fmt.Println("encode:", flag)
	Decode(flag)
	fmt.Println(flag)
	fmt.Println(string(flag))
}

func TestCompile(t *testing.T) {
	flag := []byte("yak{A_RE@LW0RLD_5TACKB@SE_VM}")
	Encode(flag)
	str := strings.Join(
		lo.Map(flag, func(v byte, _ int) string {
			return strconv.FormatInt(int64(v), 10)
		}),
		",",
	)
	code := `
b = []byte{` + str + `}
print("please input flag:")
a = get()
if len(a) > len(b){
	println("input string too long!")
	return
}

// println("raw input:", a)

func1 = (a) => {
	for i, v = range a {
		if i % 2 == 0 {
			a[i] = v ^ 0xf0
		} else {
			a[i] = v ^ 0x0f
		}
	}
	return a
}


// just function
a = func1(a) 


func2 := () => {
	for i, v = range a {
		t = i * 2
		a[i] = (^v & t) | (v & ^t)
	}
}

// closure function
func2() 

compare1 = () => {
	if len(a) != len(b) {
		return false
	}
	for i, v = range a {
		if v != b[i]{
			return false
		}
	}
	return true
}

// this function in clouse and native
if compare1() { 
	print("yes! you get it!")
}else {
	print("no this not flag")
}
	`
	// engine := yak.NewScriptEngine(100)
	// bytes, err := engine.Compile(code)

	cl := yakast.NewYakCompiler()
	cl.Compiler(code)
	SymbolTable, Codes := cl.GetRootSymbolTable(), cl.GetOpcodes()
	m := yakvm.NewCodesMarshaller()
	bytes, err := m.Marshal(SymbolTable, Codes)
	if err != nil {
		panic("marshal error")
	}

	err = os.WriteFile("./marshalCode", bytes, 0600)
	if err != nil {
		panic("write file error")
	}
}
