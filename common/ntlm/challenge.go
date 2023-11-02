package ntlm

import (
	"bytes"
	"encoding/binary"
	"github.com/yaklang/yaklang/common/utils"
)

type challengeMessageFields struct {
	messageHeader
	TargetName      varField
	NegotiateFlags  negotiateFlags //标志位
	ServerChallenge [8]byte        //challenge
	_               [8]byte        //Reserved 8 bytes 占位 发送0填充、接收时忽略
	TargetInfo      varField
}

func (m challengeMessageFields) IsValid() bool { // 检查type
	return m.messageHeader.IsValid() && m.MessageType == 2
}

type challengeMessage struct {
	challengeMessageFields
	TargetName string
	//TargetInfo    map[avID][]byte
	TargetInfoRaw []byte
}

func (m *challengeMessage) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)

	err := binary.Read(r, binary.LittleEndian, &m.challengeMessageFields) // 解析 fields
	if err != nil {
		return err
	}

	if !m.challengeMessageFields.IsValid() {
		return utils.Errorf("Message is not a valid challenge message: %+v", m.challengeMessageFields.messageHeader)
	}

	if m.challengeMessageFields.TargetName.Len > 0 { // 解析 TargetName
		m.TargetName, err = m.challengeMessageFields.TargetName.ReadStringFrom(data, m.NegotiateFlags.Has(negotiateFlagNTLMSSPNEGOTIATEUNICODE))
		if err != nil {
			return err
		}
	}

	if m.challengeMessageFields.TargetInfo.Len > 0 {

	}

	return nil
}
