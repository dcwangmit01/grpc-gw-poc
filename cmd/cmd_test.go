package cmd_test

import (
	"github.com/dcwangmit01/grpc-gw-poc/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"io"
	"os"
)

var _ = Describe("Cmd", func() {
	// The Ginkgo test runner takes over os.Args and fills it with its own
	// flags.  This makes the cobra command arg parsing fail because of
	// unexpected options.  Work around this.

	// Save a copy of os.Args
	var origArgs = os.Args[:]

	// A non-threadsafe buffer for capturing stdout
	var buf bytes.Buffer

	BeforeEach(func() {
		// Trim os.Args to only the first arg, which is the command itself
		os.Args = os.Args[:1]

		// set the output to both Stdout and a byteBuffer
		mw := io.MultiWriter(&buf, os.Stdout)
		cmd.RootCmd.SetOutput(mw)
	})

	AfterEach(func() {
		// Restore os.Args
		os.Args = origArgs[:]

		// restore the output to Stdout
		cmd.RootCmd.SetOutput(os.Stdout)
	})

	Describe("RootCmd", func() {
		Context("When run with no args", func() {
			It("Should show rootcmd help", func() {
				// Run the command which outputs to stdout
				err := cmd.RootCmd.Execute()

				// bytes.Buffer.String() returns the contents of the unread
				// portion of the buffer as a string.
				out := buf.String()

				// process the output
				//  (?s): allows for "." to represent "\n"
				Expect(out).Should(MatchRegexp("(?s)grpc-gw-poc.*help.*keyval"))
				Expect(err).Should(BeNil())
			})
		})
	})

	Describe("keyval Subcommand", func() {
		Context("When run with no args", func() {
			It("Should show keyval help", func() {
				// Set args to command
				os.Args = append(os.Args, "keyval")

				// Run the command which outputs to stdout
				err := cmd.RootCmd.Execute()

				// bytes.Buffer.String() returns the contents of the unread
				// portion of the buffer as a string.
				out := buf.String()

				// process the output
				//  (?s): allows for "." to represent "\n"
				Expect(out).Should(MatchRegexp("(?s)grpc-gw-poc.*help.*keyval.*create.*read.*update.*delete"))
				Expect(err).Should(BeNil())
			})
		})
	})
})
