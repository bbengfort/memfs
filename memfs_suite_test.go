package memfs_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const TempDirPrefix = "com.bengfort.memfs-"

func TestMemfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memfs Suite")
}

//===========================================================================
// Testing Helper Functions
//===========================================================================

// Runes for the random string function
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Create a random string of length n
func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
