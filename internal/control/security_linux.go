//go:build linux
// +build linux

package control

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

// verifyPeer performs the actual syscall to check UID/GID on Linux.
func verifyPeer(conn net.Conn) error {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return nil // Not a unix connection
	}

	raw, err := unixConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("failed to get raw connection: %w", err)
	}

	var ucred *syscall.Ucred
	var sysErr error

	err = raw.Control(func(fd uintptr) {
		ucred, sysErr = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	})

	if err != nil {
		return fmt.Errorf("failed to execute control: %w", err)
	}
	if sysErr != nil {
		return fmt.Errorf("failed to get SO_PEERCRED: %w", sysErr)
	}

	// Logic: Only root (0) or the current process owner can connect.
	if ucred.Uid != 0 && uint32(os.Geteuid()) != ucred.Uid {
		return fmt.Errorf("unauthorized UID: %d", ucred.Uid)
	}

	return nil
}
