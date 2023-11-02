package ntlm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

type negotiateFlags uint32

const (
	/*A*/ negotiateFlagNTLMSSPNEGOTIATEUNICODE negotiateFlags = 1 << 0
	/*B*/ negotiateFlagNTLMNEGOTIATEOEM = 1 << 1
	/*C*/ negotiateFlagNTLMSSPREQUESTTARGET = 1 << 2

	/*D*/
	negotiateFlagNTLMSSPNEGOTIATESIGN = 1 << 4
	/*E*/ negotiateFlagNTLMSSPNEGOTIATESEAL = 1 << 5
	/*F*/ negotiateFlagNTLMSSPNEGOTIATEDATAGRAM = 1 << 6
	/*G*/ negotiateFlagNTLMSSPNEGOTIATELMKEY = 1 << 7

	/*H*/
	negotiateFlagNTLMSSPNEGOTIATENTLM = 1 << 9

	/*J*/
	negotiateFlagANONYMOUS = 1 << 11
	/*K*/ negotiateFlagNTLMSSPNEGOTIATEOEMDOMAINSUPPLIED = 1 << 12
	/*L*/ negotiateFlagNTLMSSPNEGOTIATEOEMWORKSTATIONSUPPLIED = 1 << 13

	/*M*/
	negotiateFlagNTLMSSPNEGOTIATEALWAYSSIGN = 1 << 15
	/*N*/ negotiateFlagNTLMSSPTARGETTYPEDOMAIN = 1 << 16
	/*O*/ negotiateFlagNTLMSSPTARGETTYPESERVER = 1 << 17

	/*P*/
	negotiateFlagNTLMSSPNEGOTIATEEXTENDEDSESSIONSECURITY = 1 << 19
	/*Q*/ negotiateFlagNTLMSSPNEGOTIATEIDENTIFY = 1 << 20

	/*R*/
	negotiateFlagNTLMSSPREQUESTNONNTSESSIONKEY = 1 << 22
	/*S*/ negotiateFlagNTLMSSPNEGOTIATETARGETINFO = 1 << 23

	/*T*/
	negotiateFlagNTLMSSPNEGOTIATEVERSION = 1 << 25

	/*U*/
	negotiateFlagNTLMSSPNEGOTIATE128 = 1 << 29
	/*V*/ negotiateFlagNTLMSSPNEGOTIATEKEYEXCH = 1 << 30
	/*W*/ negotiateFlagNTLMSSPNEGOTIATE56 = 1 << 31
)

const expMsgBodyLen = 40

type NegotiateMessage struct {
	messageHeader                 // 固定头： NTLMSSP\0 01
	NegotiateFlags negotiateFlags // 标志位

	Domain      varField // security buffer
	Workstation varField

	Version // 8 byte version
}

var defaultFlags = negotiateFlagNTLMSSPNEGOTIATETARGETINFO |
	negotiateFlagNTLMSSPNEGOTIATE56 |
	negotiateFlagNTLMSSPNEGOTIATE128 |
	negotiateFlagNTLMSSPNEGOTIATEUNICODE |
	negotiateFlagNTLMSSPNEGOTIATEEXTENDEDSESSIONSECURITY

func NewNegotiateMessageDefault(domainName, workstationName string) ([]byte, error) { //默认的flags和version版本
	payloadOffset := expMsgBodyLen
	flags := defaultFlags

	if domainName != "" {
		flags |= negotiateFlagNTLMSSPNEGOTIATEOEMDOMAINSUPPLIED
	}

	if workstationName != "" {
		flags |= negotiateFlagNTLMSSPNEGOTIATEOEMWORKSTATIONSUPPLIED
	}

	msg := NegotiateMessage{
		messageHeader:  newMessageHeader(1),
		NegotiateFlags: flags,
		Domain:         newVarField(&payloadOffset, len(domainName)),
		Workstation:    newVarField(&payloadOffset, len(workstationName)),
		Version:        DefaultVersion(),
	}

	b := bytes.Buffer{}
	if err := binary.Write(&b, binary.LittleEndian, &msg); err != nil { // 小端序写入
		return nil, err
	}

	if b.Len() != expMsgBodyLen {
		return nil, errors.New("incorrect body length")
	}

	payload := strings.ToUpper(domainName + workstationName)
	if _, err := b.WriteString(payload); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (field *negotiateFlags) Has(flags negotiateFlags) bool { // 检查标志位
	return *field&flags == flags
}

func (field *negotiateFlags) Unset(flags negotiateFlags) { // 清除标志位
	*field = *field ^ (*field & flags)
}

type Version struct { // 当设置指定flag赋值,否00
	ProductMajorVersion uint8
	ProductMinorVersion uint8
	ProductBuild        uint16
	_                   [3]byte //填充数据 3 byte 00
	NTLMRevisionCurrent uint8
}

func DefaultVersion() Version {
	return Version{
		ProductMajorVersion: 6,
		ProductMinorVersion: 1,
		ProductBuild:        7601,
		NTLMRevisionCurrent: 15,
	}
}
