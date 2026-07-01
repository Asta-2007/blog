package utility

import (
	"crypto/rand"
	"log"
	"math/big"
)

func UniqueID(n int) string {
	c := []byte("1234567890")
	r := make([]byte, n)
	for i := range r {
		r[i] = c[cryptoRandSecure(int64(len(c)))]
	}

	return string(r)
}

func cryptoRandSecure(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		log.Printf("uuid crypto err : %s", err.Error())
	}
	return nBig.Int64()
}
