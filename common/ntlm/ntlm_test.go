package ntlm

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/utils"
	"testing"
)

func TestNTLMhash(t *testing.T) {
	if string(GetNtlmHash("test123")) != "c5a237b7e9d8e708d8436b6148a25fa1" {
		t.Fatalf("ntlm hash not correct: %v", string(GetNtlmHash("test123")))
	}
	GetLmHash("test123")
}

func TestToUnicode(t *testing.T) {
	a := toUnicode("test123")
	spew.Dump(a)
	b := utils.CalcMd4(a)
	spew.Dump(b)
}
