//go:build !linux
// +build !linux

package control

import "net"

// verifyPeer is a no-op on non-linux platforms.
func verifyPeer(conn net.Conn) error {
	return nil
}
