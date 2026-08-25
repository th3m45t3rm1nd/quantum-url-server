package main

import (
	"strings"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encode(n uint64) string {
	if n == 0 {
		return "0"
	}

	var out []byte

	for n > 0 {
		out = append(out, alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j+1 {
		out[i], out[j] = out[j], out[i]
	}

	return string(out)
}

func decode(str string) uint64 {
	var out int = 0
	for _, s := range str {
		pos := strings.IndexRune(alphabet, s)
		if pos == -1 {
			panic("Invalid Character")
		}

		out = out*62 + pos
	}

	return uint64(out)

}
