package simple_rpc

import (
	"encoding/binary"
	"net"
)

const lenBytes = 8

func Read(conn net.Conn) []byte {
	res := make([]byte, lenBytes)
	conn.Read(res)
	res = make([]byte, binary.BigEndian.Uint64(res))
	conn.Read(res)
	return res
}

func Write(conn net.Conn, data []byte) {
	buf := make([]byte, lenBytes+len(data))
	binary.BigEndian.PutUint64(buf[:lenBytes], uint64(len(data)))
	copy(buf[lenBytes:], data)
	conn.Write(buf)
}
