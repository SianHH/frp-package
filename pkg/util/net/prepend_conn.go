package net

import (
	"net"
	"time"
)

type PrependConn struct {
	net.Conn
	prepend []byte
	pos     int
}

func NewPrependConn(conn net.Conn, prepend []byte) net.Conn {
	return &PrependConn{
		Conn:    conn,
		prepend: prepend,
		pos:     0,
	}
}

func (c *PrependConn) Read(b []byte) (int, error) {
	if c.pos < len(c.prepend) {
		n := copy(b, c.prepend[c.pos:])
		c.pos += n
		return n, nil
	}
	return c.Conn.Read(b)
}

// 其它方法直接透传
func (c *PrependConn) Write(b []byte) (int, error)        { return c.Conn.Write(b) }
func (c *PrependConn) Close() error                       { return c.Conn.Close() }
func (c *PrependConn) LocalAddr() net.Addr                { return c.Conn.LocalAddr() }
func (c *PrependConn) RemoteAddr() net.Addr               { return c.Conn.RemoteAddr() }
func (c *PrependConn) SetDeadline(t time.Time) error      { return c.Conn.SetDeadline(t) }
func (c *PrependConn) SetReadDeadline(t time.Time) error  { return c.Conn.SetReadDeadline(t) }
func (c *PrependConn) SetWriteDeadline(t time.Time) error { return c.Conn.SetWriteDeadline(t) }
