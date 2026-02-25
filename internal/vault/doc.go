// Package vault implements the custom GopherShip Write-Ahead Log (WAL).
// It utilizes O_DIRECT and mmap patterns to bypass the Go heap during critical reflex paths.
package vault
