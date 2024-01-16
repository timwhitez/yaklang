package formatter

import (
	"testing"
)

func TestFormatter(t *testing.T) {
	formattedCode := Format(`
var a=1
var b,c
var b,c=1,2
b,c=1,2
a++
a<-1
int(aa)
a(a,b,c)
switch a{
case 1:
println(a)
case 2:
default:
aa
}
// this is a line comment 1
if true{
// this is a line comment 2
/*a*/a/*a*/=/*a*/1/*a*/if true{
a = 1// this is a line comment 2.3
			// this is a line comment 3
dump(a)// this is a line comment 4
}else{dump(a)}


}`)
	println(formattedCode)
}
