package memfs_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/bbengfort/memfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func makeTestConfig() *Config {
	return &Config{
		Name:      "testhost",
		CacheSize: 4295000000,
		Level:     "default",
		ReadOnly:  false,
		Replicas:  make([]*Replica, 0),
		Path:      "",
	}
}

var _ = Describe("Config", func() {

	var err error
	var tmpDir string

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir("", TempDirPrefix)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("should create a correct configuration", func() {
		config := makeTestConfig()
		Ω(config.Name).ShouldNot(BeZero())
		Ω(config.CacheSize).ShouldNot(BeZero())
		Ω(config.Level).ShouldNot(BeZero())
	})

	It("should be able to dump and load a config", func() {
		alpha := makeTestConfig()
		path := filepath.Join(tmpDir, "test-config.json")

		err = alpha.Dump(path)
		Ω(err).ShouldNot(HaveOccurred())

		bravo := new(Config)
		err = bravo.Load(path)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(bravo.Name).Should(Equal(alpha.Name))
		Ω(bravo.CacheSize).Should(Equal(alpha.CacheSize))
		Ω(bravo.Level).Should(Equal(alpha.Level))
		Ω(bravo.ReadOnly).Should(Equal(alpha.ReadOnly))
		Ω(bravo.Replicas).Should(Equal(alpha.Replicas))

		Ω(alpha.Path).Should(Equal(""))
		Ω(bravo.Path).Should(Equal(path))
	})

})
