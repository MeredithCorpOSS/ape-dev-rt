package clippy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/kr/text"
)

const MinLines = 4
const MaxLines = 18
const MaxLineLimit = 65
const NoteLengthLimit = MaxLines * MaxLineLimit
const longSpace = `             `

var msgTop = [2]string{
	` ___________________________________________________________________`,
	`/                                                                   \`,
}

const ClippyTemplate = `{{range $line := .TopLines}}
{{$line}}{{end}}
   __        {{index .BottomLines 0}}
  /  \       {{index .BottomLines 1}}
  |  |       {{index .BottomLines 2}}
  @  @       {{index .BottomLines 3}}
  || ||      {{index .BottomLines 4}}
  || ||   <--{{index .BottomLines 5}}
  |\_/|      \___________________________________________________________________/
  \___/
`
const lineFormat = "| %s%s |"
const PromptTemplate = "\n{{.Question}} {{.Yes}}/{{.No}}:"
const OverriddenPromptTemplate = "\n{{.Question}} {{.Yes}}\n\n"

type Clippy struct {
	inputReader  io.Reader
	outputWriter io.Writer
	styleFunc    func(string) string
}

func NewClippy(styleFunc func(string) string) *Clippy {
	input := os.Stdin
	output := os.Stdout
	return &Clippy{input, output, styleFunc}
}

func (c *Clippy) AskBoolean(note, question string,
	isTrueByDefault bool, yesOverride bool) (bool, error) {
	if len(note) > NoteLengthLimit {
		return false, fmt.Errorf("Note cannot be longer than %d (%d)",
			NoteLengthLimit, len(note))
	}
	if strings.Contains(note, "\n") {
		return false, fmt.Errorf("Note (%#v) cannot contain newlines, "+
			"it will be wrapped automatically", note)
	}

	note = splitLongWords(note)
	note = text.Wrap(note, MaxLineLimit)
	splitNote := strings.Split(note, "\n")
	if len(splitNote) > MaxLines {
		return false, fmt.Errorf(
			"Wrapped note cannot be longer than %d lines (is %d)",
			MaxLines, len(splitNote))
	}

	diff := len(splitNote) - MinLines
	topLines := []string{}
	bottomLines := []string{}
	if diff <= 0 {
		bottomLines = append(bottomLines, msgTop[0], msgTop[1])
	} else if diff == 1 {
		topLines = append(topLines, longSpace+msgTop[0])
		bottomLines = append(bottomLines, msgTop[1])
	} else {
		topLines = append(topLines, longSpace+msgTop[0])
		topLines = append(topLines, longSpace+msgTop[1])
	}

	for i := 0; i < MaxLines; i++ {
		line := ""
		if len(splitNote)-1 >= i {
			line = splitNote[i]
		}

		if len(line) <= MaxLineLimit {
			filling := fmt.Sprintf(lineFormat, line, strings.Repeat(" ", MaxLineLimit-len(line)))
			if diff > 0 && i+len(msgTop) < diff {
				topLines = append(topLines, longSpace+filling)
			} else {
				bottomLines = append(bottomLines, filling)
			}
		}
	}

	data := struct {
		TopLines          []string
		BottomLines       []string
		Question, Yes, No string
	}{
		TopLines:    topLines,
		BottomLines: bottomLines,
		Question:    question,
		Yes:         "y",
		No:          "n",
	}
	colouredClippy := ClippyTemplate
	if c.styleFunc != nil {
		colouredClippy = c.styleFunc(ClippyTemplate)
	}

	t := template.Must(template.New("clippy").Parse(colouredClippy))

	err := t.Execute(c.outputWriter, data)
	if err != nil {
		return isTrueByDefault, err
	}
	if yesOverride {
		data.Yes = "(Answer overriden with Yes)"
		data.No = ""
		t = template.Must(template.New("prompt").Parse(OverriddenPromptTemplate))
	} else {
		if isTrueByDefault {
			data.Yes = "[y]"
		} else {
			data.No = "[n]"
		}
		t = template.Must(template.New("prompt").Parse(PromptTemplate))
	}

	err = t.Execute(c.outputWriter, data)
	if err != nil {
		return isTrueByDefault, err
	}
	if yesOverride {
		return true, nil
	}
	reader := bufio.NewReader(c.inputReader)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return isTrueByDefault, err
	}
	answer = strings.TrimSpace(answer)
	if answer == "y" {
		return true, nil
	}

	return isTrueByDefault, nil
}

func splitLongWords(note string) string {
	words := strings.Split(note, " ")
	newNote := ""
	for _, word := range words {
		if len(word) < MaxLineLimit {
			newNote = newNote + " " + word
		} else {
			runes := []rune(word)
			for i, char := range runes {
				if i%MaxLineLimit == 0 {
					newNote = newNote + " " + string(char)
				} else {
					newNote = newNote + string(char)
				}
			}
		}
	}
	return newNote
}
