package ntlm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/yaklang/yaklang/common/utils"
	"strings"
	"unicode/utf16"
)

//func GetNtlmV2Hash(password, username, target string) []byte {
//	return hmacMd5(GetNtlmHash(password), toUnicode(strings.ToUpper(username)+target))
//}

func GetNtlmHash(password string) []byte { // ntlm hash
	return []byte(utils.CalcMd4(toUnicode(password)))
}

func toUnicode(s string) []byte {
	uints := utf16.Encode([]rune(s))
	b := bytes.Buffer{}
	binary.Write(&b, binary.LittleEndian, &uints)
	return b.Bytes()
}

func GetLmHash(password string) []byte {
	passHex := hex.EncodeToString([]byte(strings.ToUpper(password)))
	if len(passHex) > 28 {
		passHex = passHex[:28]
	} else {
		zeroHex := strings.Repeat("0", 28-len(passHex))
		passHex = passHex + zeroHex
	}

	return []byte(passHex)
}
