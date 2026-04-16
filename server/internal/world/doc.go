// Package world is the top-level runtime aggregate: it holds a single
// board and the set of connected players, and exposes operations like
// "apply move" and "snapshot for client". It is the source of truth at
// runtime.
package world
