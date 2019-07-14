// Package minhash contains implementations of KMV and KHF MinHash algorithms
package minhash

// MinHash is an interface to group the different flavours of MinHash implemented here
type MinHash interface {
	AddHash(uint64)
	GetSketch() []uint64
	GetSimilarity(mh2 MinHash) (float64, error)
}
