package ntlm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"strconv"
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

func GetLMHash(password string) []byte { // lm hash
	passHex := hex.EncodeToString([]byte(strings.ToUpper(password)))
	if len(passHex) > 28 {
		passHex = passHex[:28]
	} else {
		zeroHex := strings.Repeat("0", 28-len(passHex))
		passHex = passHex + zeroHex
	}

	leftStream := hex2Bin(passHex[:14])
	rightStream := hex2Bin(passHex[14:])

	leftStream = bin2Hex(regroupStream(leftStream, 7))
	rightStream = bin2Hex(regroupStream(rightStream, 7))

	leftByte, _ := hex.DecodeString(leftStream)
	rightByte, _ := hex.DecodeString(rightStream)

	leftLMHash, _ := codec.DESECBEnc(leftByte, []byte("KGS!@#$%"))
	rightLMHash, _ := codec.DESECBEnc(rightByte, []byte("KGS!@#$%"))

	return []byte(hex.EncodeToString(leftLMHash) + hex.EncodeToString(rightLMHash))
}

func hex2Bin(hexStr string) string { //hex string to binary string
	hex, _ := strconv.ParseInt(hexStr, 16, 64)
	binStr := strconv.FormatInt(hex, 2)
	binStr = strings.Repeat("0", len(hexStr)*4-len(binStr)) + binStr
	return binStr
}

func bin2Hex(binStr string) string {
	bin, _ := strconv.ParseInt(binStr, 2, 64)
	hexStr := strconv.FormatInt(bin, 16)
	hexStr = strings.Repeat("0", len(binStr)/4-len(hexStr)) + hexStr
	return hexStr
}

func regroupStream(binStream string, step int) string {
	var newStream string
	for i := 0; i < len(binStream); i += step {
		if i+step > len(binStream) {
			newStream += binStream[i:] + strings.Repeat("0", i+step-len(binStream))
			break
		}
		newStream += binStream[i:i+step] + "0"
	}
	return newStream
}
