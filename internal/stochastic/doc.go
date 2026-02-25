// Package stochastic provides lazy, atomic environmental awareness for GopherShip.
// It allows performance-critical modules to check global system state (pressure zones)
// without incurring the cost of constant cache-line contention or heavy locking.
package stochastic
