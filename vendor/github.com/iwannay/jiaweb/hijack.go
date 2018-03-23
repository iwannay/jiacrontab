package jiaweb

import (
	"bufio"
	"net"
)

type HijackConn struct {
	ReadWriter *bufio.ReadWriter
	Conn       net.Conn
	header     string
}

func (h *HijackConn) WriteString(content string) (int, error) {
	n, err := h.ReadWriter.WriteString(h.header + "\r\n" + content)
	if err == nil {
		h.ReadWriter.Flush()
	}
	return n, err
}

func (h *HijackConn) WriteBlob(p []byte) (size int, err error) {
	size, err = h.ReadWriter.Write(p)
	if err == nil {
		h.ReadWriter.Flush()
	}
	return
}

func (h *HijackConn) SetHedader(key, value string) {
	h.header += key + ": " + value + "\r\n"
}

func (h *HijackConn) Close() error {
	return h.Conn.Close()
}
