//go:build !linux && !windows
// +build !linux,!windows

package dokodemo

import (
	"fmt"
	"net"
)

func FakeUDP(conn net.Conn, addr *net.UDPAddr, mark int) (net.PacketConn, error) {
	return nil, &net.OpError{Op: "fake", Err: fmt.Errorf("!linux")}
}
