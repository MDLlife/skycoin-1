package cipher

import (
	"encoding/hex"
	"errors"
	"hash"
	"log"
)

var (
	// Memory pool for hashes
	sha256HashPoolSize    = 30
	sha256HashPool        chan hash.Hash
	ripemd160HashPoolSize = 30
	ripemd160HashPool     chan hash.Hash
)

var (
	// ErrInvalidLengthRipemd160 Invalid ripemd160 length
	ErrInvalidLengthRipemd160 = errors.New("Invalid ripemd160 length")
	// ErrInvalidLengthSHA256    Invalid sha256 length
	ErrInvalidLengthSHA256 = errors.New("Invalid sha256 length")
	// ErrInvalidHexLength       Invalid hex length
	ErrInvalidHexLength = errors.New("Invalid hex length")
	// ErrInvalidBytesLength     Invalid bytes length
	ErrInvalidBytesLength = errors.New("Invalid bytes length")
)

// Ripemd160 ripemd160
type Ripemd160 [20]byte

// MustSet sets value, panics on error
func (rd *Ripemd160) MustSet(b []byte) {
	if len(b) != 20 {
		err := NewValueError(
			ErrInvalidLengthRipemd160, "b", b,
		)
		log.Print(err)
		panic(err)
	}
	copy(rd[:], b[:])
}

// Set sets value
func (rd *Ripemd160) Set(b []byte) error {
	if len(b) != 20 {
		return errors.New("Invalid ripemd160 length")
	}
	copy(rd[:], b[:])
	return nil
}

// HashRipemd160 hash data to Ripemd160
func HashRipemd160(data []byte) Ripemd160 {
	ripemd160Hash := <-ripemd160HashPool
	ripemd160Hash.Reset()
	// ripemd160.Write never returns an error
	ripemd160Hash.Write(data) // nolint: errcheck
	sum := ripemd160Hash.Sum(nil)
	ripemd160HashPool <- ripemd160Hash

	h := Ripemd160{}
	h.MustSet(sum)
	return h
}

// SHA256 32 bytes
type SHA256 [32]byte

// MustSet sets value, panics on error
func (g *SHA256) MustSet(b []byte) {
	if len(b) != 32 {
		err := NewValueError(
			ErrInvalidLengthSHA256, "b", b,
		)
		log.Print(err)
		panic(err)
	}
	copy(g[:], b[:])
}

// Set sets value
func (g *SHA256) Set(b []byte) error {
	if len(b) != 32 {
		return errors.New("Invalid sha256 length")
	}
	copy(g[:], b[:])
	return nil
}

// Hex encode sha256 to hex string
func (g SHA256) Hex() string {
	return hex.EncodeToString(g[:])
}

// Null returns true if the hash is null (0x0000..)
func (g SHA256) Null() bool {
	return g == SHA256{}
}

// Xor xor
func (g *SHA256) Xor(b SHA256) SHA256 {
	c := SHA256{}
	for i := 0; i < 32; i++ {
		c[i] = g[i] ^ b[i]
	}
	return c
}

// SumSHA256 sum sha256
func SumSHA256(b []byte) SHA256 {
	sha256Hash := <-sha256HashPool
	sha256Hash.Reset()
	// sha256.Write never returns an error
	sha256Hash.Write(b) // nolint: errcheck
	sum := sha256Hash.Sum(nil)
	sha256HashPool <- sha256Hash

	h := SHA256{}
	h.MustSet(sum)
	return h
}

// SHA256FromHex decodes a hex encoded SHA256 hash to bytes.  If invalid, will return error.
// Does not panic.
func SHA256FromHex(hs string) (SHA256, error) {
	h := SHA256{}
	b, err := hex.DecodeString(hs)
	if err != nil {
		return h, err
	}
	if len(b) != len(h) {
		return h, ErrInvalidHexLength
	}
	h.MustSet(b)
	return h, nil
}

// MustSHA256FromHex same as SHA256FromHex, except will panic when detect error
func MustSHA256FromHex(hs string) SHA256 {
	h, err := SHA256FromHex(hs)
	if err != nil {
		log.Panic(err)
	}
	return h
}

// SHA256FromBytes converts []byte to SHA256
func SHA256FromBytes(b []byte) (SHA256, error) {
	h := SHA256{}
	err := h.Set(b)
	return h, err
}

// MustSHA256FromBytes is the same as SHA256FromBytes, except it will panic when it detects an error
func MustSHA256FromBytes(b []byte) SHA256 {
	h, err := SHA256FromBytes(b)
	if err != nil {
		log.Panic(err)
	}
	return h
}

// MustSumSHA256 like SumSHA256, but len(b) must equal n, or panic
func MustSumSHA256(b []byte, n int) SHA256 {
	if len(b) != n {
		log.Panicf("Invalid sumsha256 byte length. Expected %d, have %d", n, len(b))
	}
	return SumSHA256(b)
}

// DoubleSHA256 double SHA256
func DoubleSHA256(b []byte) SHA256 {
	h1 := SumSHA256(b)
	h2 := SumSHA256(h1[:])
	return h2
}

// AddSHA256 returns the SHA256 hash of to two concatenated hashes
func AddSHA256(a SHA256, b SHA256) SHA256 {
	c := append(a[:], b[:]...)
	return SumSHA256(c)
}

// Returns the next highest power of 2 above n, if n is not already a
// power of 2
func nextPowerOfTwo(n uint64) uint64 {
	var k uint64 = 1
	for k < n {
		k *= 2
	}
	return k
}

// Merkle computes the merkle root of a hash array
// Array of hashes is padded with 0 hashes until next power of 2
func Merkle(h0 []SHA256) SHA256 {
	lh := uint64(len(h0))
	np := nextPowerOfTwo(lh)
	h1 := append(h0, make([]SHA256, np-lh)...)
	for len(h1) != 1 {
		h2 := make([]SHA256, len(h1)/2)
		for i := 0; i < len(h2); i++ {
			h2[i] = AddSHA256(h1[2*i], h1[2*i+1])
		}
		h1 = h2
	}
	return h1[0]
}
