package clippy

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestAskBoolean_basic(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	checkBooleanConversation("y\n", true, c, w, t)
	checkBooleanConversation("yes\n", false, c, w, t)
	checkBooleanConversation("n\n", false, c, w, t)
	checkBooleanConversation("\n", false, c, w, t)
	checkBooleanConversation("yada\n", false, c, w, t)
}

func TestAskBoolean_defaultAnswer(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	checkSomeEndingLineOfBooleanConversation(
		"Yo?", true, "Yo? [y]/n:", c, w, output, t, false, 1)
	checkSomeEndingLineOfBooleanConversation(
		"Yo?", false, "Yo? y/[n]:", c, w, output, t, false, 1)
}

func TestAskBoolean_yesOverride(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	checkSomeEndingLineOfBooleanConversation(
		"Yo?", true, "Yo? (Answer overriden with Yes)", c, w, output, t, true, 3)
}

func TestAskBoolean_3_lines_output(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
   __         ___________________________________________________________________
  /  \       /                                                                   \
  |  |       | Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do   |
  @  @       | eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut    |
  || ||      | enim ad minim veniam, quis nostrud exercitation ullamco           |
  || ||   <--|                                                                   |
  |\_/|      \___________________________________________________________________/
  \___/

Yo? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit,"+
			" sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."+
			" Ut enim ad minim veniam, quis nostrud exercitation ullamco",
		"Yo?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func TestAskBoolean_2_lines_output(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
   __         ___________________________________________________________________
  /  \       /                                                                   \
  |  |       | Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus  |
  @  @       | est eros yada.                                                    |
  || ||      |                                                                   |
  || ||   <--|                                                                   |
  |\_/|      \___________________________________________________________________/
  \___/

Yo? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus "+
			"est eros yada.", "Yo?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func TestAskBoolean_1_line_output(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
   __         ___________________________________________________________________
  /  \       /                                                                   \
  |  |       | Yada yada yada...                                                 |
  @  @       |                                                                   |
  || ||      |                                                                   |
  || ||   <--|                                                                   |
  |\_/|      \___________________________________________________________________/
  \___/

Yo? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"Yada yada yada...", "Yo?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func TestAskBoolean_edgeCaseLineLength(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
   __         ___________________________________________________________________
  /  \       /                                                                   \
  |  |       | Application "example_standalone" doesn't exist in "test", do you  |
  @  @       | want to create it?                                                |
  || ||      |                                                                   |
  || ||   <--|                                                                   |
  |\_/|      \___________________________________________________________________/
  \___/

continue? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"Application \"example_standalone\" doesn't exist in \"test\", do you want to create it?", "continue?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func checkBooleanConversation(input string, expectedAnswer bool,
	c *Clippy, w io.Writer, t *testing.T) {
	go w.Write([]byte(input))

	answer, err := c.AskBoolean("Hmm...", "Yo?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer to %#v was %t, given %t",
			input, expectedAnswer, answer)
	}
}

func checkSomeEndingLineOfBooleanConversation(input string, defaultAnswer bool,
	expectedLastLine string, c *Clippy, w io.Writer, o *bytes.Buffer, t *testing.T, y bool, offset int) {
	go w.Write([]byte("\n"))

	_, err := c.AskBoolean("", input, defaultAnswer, y)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	lines := strings.Split(o.String(), "\n")
	lastLine := lines[len(lines)-offset]
	if lastLine != expectedLastLine {
		t.Errorf("Expected last line: %#v for default answer %t, given:\n %s",
			expectedLastLine, defaultAnswer, lastLine)
	}
}

func TestLongWordTextWrapNoChange(t *testing.T) {

	expectedOutput := ` All of these words are under the word limit`

	output := splitLongWords(`All of these words are under the word limit`)

	if expectedOutput != output {
		t.Errorf("Expected output: %s\ngiven output: %s",
			expectedOutput, output)
	}
}

func TestLongWordTextWrap(t *testing.T) {

	expectedOutput := ` All of these words are under the word limit except this one: ` +
		`12345678901234567890123456789012345678901234567890123456789012345 67890`

	output := splitLongWords(`All of these words are under the word limit except this one: ` +
		`1234567890123456789012345678901234567890123456789012345678901234567890`)

	if expectedOutput != output {
		t.Errorf("Expected output: %s\ngiven output: %s",
			expectedOutput, output)
	}
}

func TestLongWordTextWrap4Lines(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
   __         ___________________________________________________________________
  /  \       /                                                                   \
  |  |       | 12345678901234567890123456789012345678901234567890123456789012345 |
  @  @       | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||      | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||   <--| 12345678901234567890123456789012345678901234567890123456789012345 |
  |\_/|      \___________________________________________________________________/
  \___/

continue? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345", "continue?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func Test5LinesOutput(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
              ___________________________________________________________________
   __        /                                                                   \
  /  \       | 12345678901234567890123456789012345678901234567890123456789012345 |
  |  |       | 12345678901234567890123456789012345678901234567890123456789012345 |
  @  @       | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||      | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||   <--| 12345678901234567890123456789012345678901234567890123456789012345 |
  |\_/|      \___________________________________________________________________/
  \___/

continue? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345",
		"continue?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func Test6LinesOutput(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
              ___________________________________________________________________
             /                                                                   \
   __        | 12345678901234567890123456789012345678901234567890123456789012345 |
  /  \       | 12345678901234567890123456789012345678901234567890123456789012345 |
  |  |       | 12345678901234567890123456789012345678901234567890123456789012345 |
  @  @       | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||      | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||   <--| 12345678901234567890123456789012345678901234567890123456789012345 |
  |\_/|      \___________________________________________________________________/
  \___/

continue? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345",
		"continue?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}

func Test18LinesOutput(t *testing.T) {
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	output := new(bytes.Buffer)
	c := &Clippy{
		inputReader:  r,
		outputWriter: output,
	}

	expectedOutput := `
              ___________________________________________________________________
             /                                                                   \
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
             | 12345678901234567890123456789012345678901234567890123456789012345 |
   __        | 12345678901234567890123456789012345678901234567890123456789012345 |
  /  \       | 12345678901234567890123456789012345678901234567890123456789012345 |
  |  |       | 12345678901234567890123456789012345678901234567890123456789012345 |
  @  @       | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||      | 12345678901234567890123456789012345678901234567890123456789012345 |
  || ||   <--| 12345678901234567890123456789012345678901234567890123456789012345 |
  |\_/|      \___________________________________________________________________/
  \___/

continue? y/[n]:`

	go w.Write([]byte("\n"))

	_, err := c.AskBoolean(
		"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345"+
			"12345678901234567890123456789012345678901234567890123456789012345",
		"continue?", false, false)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if expectedOutput != output.String() {
		t.Errorf("Expected output: %s\ngiven output %s",
			expectedOutput, output.String())
	}
}
