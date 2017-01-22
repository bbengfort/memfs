package memfs_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/bbengfort/memfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {

	Describe("log level handling", func() {

		It("should be able to convert a log level to a string", func() {
			Ω(LevelDebug.String()).Should(Equal("DEBUG"))
			Ω(LevelInfo.String()).Should(Equal("INFO"))
			Ω(LevelWarn.String()).Should(Equal("WARN"))
			Ω(LevelError.String()).Should(Equal("ERROR"))
			Ω(LevelFatal.String()).Should(Equal("FATAL"))
		})

		It("should be able to convert a string to a log level", func() {
			Ω(LevelFromString("DEBUG")).Should(Equal(LevelDebug))
			Ω(LevelFromString("INFO")).Should(Equal(LevelInfo))
			Ω(LevelFromString("WARN")).Should(Equal(LevelWarn))
			Ω(LevelFromString("WARNING")).Should(Equal(LevelWarn))
			Ω(LevelFromString("ERROR")).Should(Equal(LevelError))
			Ω(LevelFromString("FATAL")).Should(Equal(LevelFatal))
		})

		It("should convert any case string to a log level", func() {
			Ω(LevelFromString("INFO")).Should(Equal(LevelInfo))
			Ω(LevelFromString("info")).Should(Equal(LevelInfo))
			Ω(LevelFromString("Info")).Should(Equal(LevelInfo))
			Ω(LevelFromString("InFo")).Should(Equal(LevelInfo))
		})

		It("should convert strings with spaces to a log level", func() {
			Ω(LevelFromString("ERROR ")).Should(Equal(LevelError))
			Ω(LevelFromString("ERROR   ")).Should(Equal(LevelError))
			Ω(LevelFromString("   ERROR")).Should(Equal(LevelError))
			Ω(LevelFromString("   ERROR   ")).Should(Equal(LevelError))
		})

	})

	Describe("logging methods", func() {

		var (
			err     error   // captured errors
			testDir string  // path to temporary test files
			logger  *Logger // logger instantiated from config
		)

		Context("to stdout", func() {

			BeforeEach(func() {
				logger, err = InitLogger("", "INFO")

				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			})

			It("should output to os.Stdout", func() {
				Ω(logger.GetHandler()).Should(Equal(os.Stdout))
			})

			It("should be able to set a new io.Writer for output", func() {

				// Temporary Buffer
				type Buffer struct {
					bytes.Buffer
					io.Closer
				}

				// Reset the logger handler
				buf := new(Buffer)
				logger.SetHandler(buf)

				// Write a log message and test the output
				logger.Log("test log message", LevelInfo)
				logPattern := `INFO    \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: test log message\n`
				Ω(buf.String()).Should(MatchRegexp(logPattern))
			})

		})

		Context("to a log file", func() {

			var path string

			BeforeEach(func() {
				// Create a temporary log directory
				testDir, err = ioutil.TempDir("", TempDirPrefix)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				// Create a config with the temporary path
				path = filepath.Join(testDir, "testing.log")

				// Initialize the logger
				Ω(path).ShouldNot(BeAnExistingFile())
				logger, err = InitLogger(path, "INFO")
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			})

			AfterEach(func() {
				// Delete the temporary log file
				err = os.RemoveAll(testDir)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				Ω(path).ShouldNot(BeAnExistingFile())
			})

			It("should output to an open file", func() {
				Ω(path).Should(BeAnExistingFile())
			})

			It("should not log below the log level", func() {
				Ω(logger.Level).Should(Equal(LevelInfo))
				logger.Log("should not be logged", LevelDebug)
				logger.Log("should be logged", LevelInfo)
				logger.Log("definitely should be logged", LevelWarn)

				data, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				lines := strings.Split(string(data), "\n")
				Ω(lines).Should(HaveLen(3))

				Ω(lines[0]).Should(MatchRegexp(`INFO    \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: should be logged`))
				Ω(lines[1]).Should(MatchRegexp(`WARN    \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: definitely should be logged`))
				Ω(lines[2]).Should(BeZero())

			})

			It("should log debug messages", func() {
				logger.Level = LevelDebug
				logger.Debug("this is a debug message")

				data, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				Ω(string(data)).Should(MatchRegexp(`DEBUG   \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: this is a debug message`))
			})

			It("should log info messages", func() {
				logger.Info("for your information")

				data, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				Ω(string(data)).Should(MatchRegexp(`INFO    \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: for your information`))
			})

			It("should log warning messages", func() {
				logger.Warn("be careful!")

				data, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				Ω(string(data)).Should(MatchRegexp(`WARN    \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: be careful!`))
			})

			It("should log error messages", func() {
				logger.Error("there was a problem!")

				data, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

				Ω(string(data)).Should(MatchRegexp(`ERROR   \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}\]: there was a problem!`))
			})

			It("should log fatal messages and exit", func() {
				Skip("not sure how to check if fatal occurs")
			})
		})

	})

})
