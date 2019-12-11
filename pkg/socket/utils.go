package socket

import "encoding/binary"

func Serialize(payload []byte) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(payload)))
	return append(bs, payload...)
}

func Parse(data []byte) (len int, payload []byte) {
	//todo.简单实现，不考虑数据会很长
	return int(binary.LittleEndian.Uint32(data[:4])), data[4:]
}
