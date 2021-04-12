package ui

import (
	"bufio"
	"bytes"
	"github.com/briandowns/spinner"
	"io"
	"time"
)

// A spinner which changes speed based on writes to io.WriteCloser
func FlowSpinner(prefix string) io.WriteCloser {
	rateBuffer := new(bytes.Buffer)
	bufRate := make(chan int)
	milliSamplesPerSecond := 600

	s := spinner.New(spinner.CharSets[31], 1000*time.Millisecond)
	s.Suffix = "       \r"

	pr, pw := io.Pipe()

	go func(reader io.Reader) { //read into buffer
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			scanLine := scanner.Text()
			rateBuffer.Write([]byte(scanLine))
		}
		close(bufRate)
		s.Stop()
	}(pr)

	go func() { //sample buffer size
		sleepFor := time.Duration(500) * time.Millisecond
		for {
			bufRate <- rateBuffer.Len() * (1000 / milliSamplesPerSecond)
			rateBuffer.Reset()
			time.Sleep(sleepFor)
		}
	}()

	s.Prefix = "   " + prefix + "  "
	go func() { // consume buffer size-samples
		var ok bool
		var size int
		for ok = true; ok == true; {
			size, ok = <-bufRate
			rate := rateForBytes(size)
			s.UpdateSpeed(durationForRate(rate))
		}
	}()

	s.Start()
	return pw
}

// arbitrary values
func rateForBytes(c int) int {
	rate := 0
	if c > 100 {
		rate = 3
	} else if c > 20 {
		rate = 2
	} else if c > 0 {
		rate = 1
	} else {
		rate = 0
	}
	return rate
}

func durationForRate(rate int) time.Duration {
	d := map[int]string{
		0: "1000ms",
		1: "500ms",
		2: "300ms",
		3: "100ms",
	}[rate]
	duration, _ := time.ParseDuration(d)
	return duration
}
