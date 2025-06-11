//go:build windows
// +build windows

package dokodemo

import (
	"encoding/binary"
	"fmt"
	snet "net"
	"unsafe"

	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
	"golang.org/x/sys/windows"
)

type fakeUDPConn struct {
	*net.UDPConn
	srcIP *net.UDPAddr
}

func (c *fakeUDPConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if c.UDPConn == nil {
		return 0, snet.ErrClosed
	}

	b := make([]byte, 128)
	oob := buf.FromBytes(b)
	oob.Clear()

	if family := net.ParseAddress(addr.String()).Family(); family == net.AddressFamilyIPv4 || family == net.AddressFamilyDomain { // IPv4
		pktinf := new(windows.IN_PKTINFO)
		pktinf.Addr = [4]byte(addr.(*net.UDPAddr).IP.To4())

		cm := new(windows.WSACMSGHDR)
		cm.Level = windows.IPPROTO_IP
		cm.Type = windows.IP_PKTINFO
		lenCMSG := make([]byte, unsafe.Sizeof(cm.Len))
		binary.Encode(lenCMSG, binary.LittleEndian, uint16(unsafe.Sizeof(*cm)+unsafe.Sizeof(*pktinf)))
		binary.Write(oob, binary.LittleEndian, lenCMSG)
		binary.Write(oob, binary.LittleEndian, cm.Level)
		binary.Write(oob, binary.LittleEndian, cm.Type)
		binary.Write(oob, binary.LittleEndian, pktinf)
	} else { // IPv6
		pktinf := new(windows.IN6_PKTINFO)
		pktinf.Addr = [16]byte(addr.(*net.UDPAddr).IP.To16())

		cm := new(windows.WSACMSGHDR)
		cm.Level = windows.IPPROTO_IPV6
		cm.Type = windows.IPV6_PKTINFO
		lenCMSG := make([]byte, unsafe.Sizeof(cm.Len))
		binary.Encode(lenCMSG, binary.LittleEndian, uint16(unsafe.Sizeof(*cm)+unsafe.Sizeof(*pktinf)))
		binary.Write(oob, binary.LittleEndian, lenCMSG)
		binary.Write(oob, binary.LittleEndian, cm.Level)
		binary.Write(oob, binary.LittleEndian, cm.Type)
		binary.Write(oob, binary.LittleEndian, pktinf)
	}

	n, _, err := c.UDPConn.WriteMsgUDP(p, oob.Bytes(), addr.(*net.UDPAddr))
	if err == nil {
		return n, err
	}

	return c.UDPConn.WriteTo(p, addr)
}

func (c *fakeUDPConn) Close() error {
	c.UDPConn = nil
	return nil
}

func FakeUDP(conn net.Conn, addr *net.UDPAddr, mark int) (net.PacketConn, error) {
	switch c := conn.(type) {
	case *net.UDPConn:
		return &fakeUDPConn{
			UDPConn: conn.(*net.UDPConn),
			srcIP:   addr,
		}, nil
	case *fakeUDPConn:
		newC := new(fakeUDPConn)
		*newC = *c
		newC.srcIP = addr
		return newC, nil
	}
	return nil, &snet.OpError{Op: "fake", Err: fmt.Errorf("unknown type of conn")}
}
