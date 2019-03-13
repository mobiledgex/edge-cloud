package util

import "math/rand"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandAscii(n int) []byte {
	output := make([]byte, n)
	random := make([]byte, n)
	_, err := rand.Read(random)
	l := len(letterBytes)
	if err == nil {
		for ii := 0; ii < n; ii++ {
			randPos := uint8(random[ii]) % uint8(l)
			output[ii] = letterBytes[randPos]
		}
	} else {
		// slower
		for ii := 0; ii < n; ii++ {
			randPos := uint8(rand.Int63()) % uint8(l)
			output[ii] = letterBytes[randPos]
		}
	}
	return output
}
