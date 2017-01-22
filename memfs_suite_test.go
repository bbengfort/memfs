package memfs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const TempDirPrefix = "com.bengfort.memfs-"

func TestMemfs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memfs Suite")
}
