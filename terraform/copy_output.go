package terraform

import (
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/prefixedio"
)

const (
	ErrorPrefix  = "e:"
	OutputPrefix = "o:"
)

// copyOutput uses output prefixes to determine whether data on stdout
// should go to stdout or stderr. This is due to panicwrap using stderr
// as the log and error channel.
func copyOutput(r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)

	pr, err := prefixedio.NewReader(r)
	if err != nil {
		panic(err)
	}

	stderrR, err := pr.Prefix(ErrorPrefix)
	if err != nil {
		panic(err)
	}
	stdoutR, err := pr.Prefix(OutputPrefix)
	if err != nil {
		panic(err)
	}
	defaultR, err := pr.Prefix("")
	if err != nil {
		panic(err)
	}

	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	if runtime.GOOS == "windows" {
		stdout = colorable.NewColorableStdout()
		stderr = colorable.NewColorableStderr()

		// colorable is not concurrency-safe when stdout and stderr are the
		// same console, so we need to add some synchronization to ensure that
		// we can't be concurrently writing to both stderr and stdout at
		// once, or else we get intermingled writes that create gibberish
		// in the console.
		wrapped := synchronizedWriters(stdout, stderr)
		stdout = wrapped[0]
		stderr = wrapped[1]
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		io.Copy(stderr, stderrR)
	}()
	go func() {
		defer wg.Done()
		io.Copy(stdout, stdoutR)
	}()
	go func() {
		defer wg.Done()
		io.Copy(stdout, defaultR)
	}()

	wg.Wait()
}
