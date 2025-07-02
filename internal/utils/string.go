package utils

import (
	"math/rand"
	"time"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandStringUsingMathRand(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	result := make([]rune, n)
	for i := 0; i < n; i++ {
		result[i] = letters[randSource.Intn(len(letters))]
	}
	return string(result)
}
