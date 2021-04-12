package ui

import (
	"bufio"
	"io"
	"time"
)

func freshLine(length int, prefix []byte, w io.Writer) int {
	for length > 0 {
		w.Write([]byte(" "))
		length--
	}
	w.Write([]byte("\r"))
	w.Write(prefix)
	return len(prefix)
}

func GitProgressPipe(op string, w io.Writer) io.WriteCloser {
	prefix := []byte("[git " + op + "] ")
	pr, pw := io.Pipe()
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		lineLength := 0
		started := time.Now()
		for scanner.Scan() {
			if time.Since(started).Seconds() > 1.5 {
				scanByte := scanner.Text()
				if scanByte == "\n" {
					w.Write([]byte("\r"))
					lineLength = freshLine(lineLength, prefix, w)
				} else if scanByte == "\r" {
					w.Write([]byte("\r"))
					lineLength = freshLine(0, prefix, w)
				} else {
					w.Write([]byte(scanByte))
					lineLength++
				}
			}
		}
		freshLine(lineLength, []byte(""), w)
		w.Write([]byte("\r"))
	}(pr)
	return pw
}
