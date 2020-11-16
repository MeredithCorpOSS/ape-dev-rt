package ui

import (
	"bytes"
	"testing"
)

func TestStreamedUi_Error(t *testing.T) {
	outErr := new(bytes.Buffer)
	ui := &StreamedUi{
		ErrorWriter: outErr,
	}
	ui.Error("HELLO")
	bufErr := ui.ErrorBuffer

	if outErr.String() != "HELLO" {
		t.Fatalf("bad outErr: %s", outErr.String())
	}

	if bufErr.String() != "HELLO" {
		t.Fatalf("bad bufErr: %s", bufErr.String())
	}

}

func TestStreamedUi_Output(t *testing.T) {
	outWriter := new(bytes.Buffer)
	ui := &StreamedUi{
		OutputWriter: outWriter,
	}
	ui.Output("HELLO")
	bufOut := ui.OutputBuffer
	if bufOut.String() != "HELLO" {
		t.Fatalf("bad: %s", bufOut.String())
	}
	if outWriter.String() != "HELLO" {
		t.Fatalf("bad: %s", outWriter.String())
	}
}
