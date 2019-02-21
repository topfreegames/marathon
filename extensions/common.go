package extensions

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func IdGenerator(size int) string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, size)
	for i := range id {
		n := rand.Int() % len(charset)
		id[i] = charset[n]
	}
	return string(id)
}
