package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

type StreamedUi struct {
	OutputWriter io.Writer
	OutputBuffer *bytes.Buffer

	ErrorWriter io.Writer
	ErrorBuffer *bytes.Buffer

	once sync.Once
}

func (u *StreamedUi) Ask(query string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

func (u *StreamedUi) AskSecret(query string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

func (u *StreamedUi) Error(message string) {
	u.once.Do(u.init)

	fmt.Fprint(u.ErrorWriter, message)
	fmt.Fprint(u.ErrorWriter, "\n")
}

func (u *StreamedUi) Info(message string) {
	u.Output(message)
}

func (u *StreamedUi) Output(message string) {
	u.once.Do(u.init)

	fmt.Fprint(u.OutputWriter, message)
	fmt.Fprint(u.OutputWriter, "\n")
}

func (u *StreamedUi) Warn(message string) {
	u.once.Do(u.init)

	fmt.Fprint(u.ErrorWriter, message)
	fmt.Fprint(u.ErrorWriter, "\n")
}

func (u *StreamedUi) Init() {
	u.once.Do(u.init)
}

func (u *StreamedUi) FlushBuffers() {
	u.once.Do(u.init)
	u.OutputBuffer.Reset()
	u.ErrorBuffer.Reset()
}

func (u *StreamedUi) ReplaceOutputWriter(w io.Writer) {
	u.once.Do(u.init)
	u.OutputBuffer.Reset()
	u.OutputWriter = wrapWriterWithBuffer(w, u.OutputBuffer)
}

func (u *StreamedUi) ReplaceErrorWriter(w io.Writer) {
	u.once.Do(u.init)
	u.ErrorBuffer.Reset()
	u.ErrorWriter = wrapWriterWithBuffer(w, u.ErrorBuffer)
}

func (u *StreamedUi) attachBufferedReadCloser(r io.ReadCloser) (io.Writer, *bytes.Buffer, chan bool) {
	pr, pw := io.Pipe()
	var outputBuffer = new(bytes.Buffer)
	buffWriter, doneChan := wrapWriteCloserWithBuffer(pw, outputBuffer)
	attachReaderToWriteCloser(r, buffWriter)
	outWriter := attachReaderToWriter(pr, u.OutputWriter)
	return outWriter, outputBuffer, doneChan
}

func (u *StreamedUi) AttachOutputReadCloser(r io.ReadCloser) (*bytes.Buffer, chan bool) {
	u.once.Do(u.init)
	outWriter, outBuffer, doneChan := u.attachBufferedReadCloser(r)
	u.OutputWriter = outWriter
	return outBuffer, doneChan
}

func (u *StreamedUi) AttachErrorReadCloser(r io.ReadCloser) (*bytes.Buffer, chan bool) {
	u.once.Do(u.init)
	errWriter, errBuffer, doneChan := u.attachBufferedReadCloser(r)
	u.ErrorWriter = errWriter
	return errBuffer, doneChan
}

func attachReaderToWriteCloser(r io.Reader, w io.WriteCloser) {
	tr := io.TeeReader(r, w)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
		}
		w.Close()
	}(tr)
}

func attachReaderToWriter(r io.Reader, w io.Writer) io.Writer {
	tr := io.TeeReader(r, w)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
		}
	}(tr)
	return w
}

func wrapWriteCloserWithBuffer(w io.WriteCloser, b *bytes.Buffer) (io.WriteCloser, chan bool) {
	doneChan := make(chan bool)
	pr, pw := io.Pipe()
	tr := io.TeeReader(pr, w)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			writeBytes := []byte(scanner.Text())
			b.Grow(len(writeBytes))
			b.Write(writeBytes)
		}
		w.Close()
		close(doneChan)
	}(tr)
	return pw, doneChan
}

func wrapWriterWithBuffer(w io.Writer, b *bytes.Buffer) io.Writer {
	pr, pw := io.Pipe()
	tr := io.TeeReader(pr, w)
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			writeBytes := []byte(scanner.Text())
			b.Grow(len(writeBytes))
			b.Write(writeBytes)
		}
	}(tr)
	return pw
}

func (u *StreamedUi) init() {
	u.ErrorBuffer = new(bytes.Buffer)
	if u.ErrorWriter == nil {
		u.ErrorWriter = os.Stderr
	}
	u.ErrorWriter = wrapWriterWithBuffer(u.ErrorWriter, u.ErrorBuffer)

	u.OutputBuffer = new(bytes.Buffer)
	if u.OutputWriter == nil {
		u.OutputWriter = os.Stdout
	}
	u.OutputWriter = wrapWriterWithBuffer(u.OutputWriter, u.OutputBuffer)
}
