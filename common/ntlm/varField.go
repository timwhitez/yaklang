package ntlm

import (
	"errors"
)

type varField struct { //security Buffer
	Len          uint16 //缓冲区大小
	MaxLen       uint16 //缓冲区大小
	BufferOffset uint32 //缓冲区指针\偏移
}

func (f varField) ReadFrom(buffer []byte) ([]byte, error) { //按偏移读取
	if len(buffer) < int(f.BufferOffset+uint32(f.Len)) {
		return nil, errors.New("Error reading data, varField extends beyond buffer")
	}
	return buffer[f.BufferOffset : f.BufferOffset+uint32(f.Len)], nil
}

func (f varField) ReadStringFrom(buffer []byte, unicode bool) (string, error) {
	d, err := f.ReadFrom(buffer)
	if err != nil {
		return "", err
	}
	if unicode {
		return fromUnicode(d)
	}
	return string(d), err
}

func newVarField(ptr *int, fieldsize int) varField {
	f := varField{
		Len:          uint16(fieldsize),
		MaxLen:       uint16(fieldsize),
		BufferOffset: uint32(*ptr),
	}
	*ptr += fieldsize
	return f
}
