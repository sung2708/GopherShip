// Package buffer provides zero-allocation binary buffer pools for GopherShip hot paths.
// It leverages sync.Pool to minimize GC pressure during high-throughput ingestion.
//
// TODO: Implement central Homeostasis registry to track cross-component memory budgets.
package buffer
