package clippy

import (
	"fmt"
	"log"

	"github.com/ttacon/chalk"
)

const PromptQuestion = "continue?"
const NegativeAnswer = "\nOK, bailing out.\n¯\\_(ツ)_/¯"
const IssuesURL = "https://github.com/TimeInc/ape-dev-rt/issues/new"

type pCallback func() (interface{}, error)

func BoolPrompt(promptNote string, yesOverride, isSensitive bool, yesCallback pCallback, noCallback pCallback) (interface{}, bool, error) {
	var style func(string) string
	style = nil
	if isSensitive {
		style = chalk.White.NewStyle().WithTextStyle(chalk.Bold).WithBackground(chalk.Red).Style
	}

	log.Printf("[DEBUG] Prompting user to confirm")
	c := NewClippy(style)
	shouldContinue, err := c.AskBoolean(promptNote, PromptQuestion, false, yesOverride)
	if err != nil {
		log.Printf("[ERROR] Received error from Clippy prompt: %s", err)
		return false, false, err
	}
	log.Printf("[DEBUG] User replied %t in the prompt", shouldContinue)

	if shouldContinue {
		out, err := yesCallback()
		return out, shouldContinue, err
	}

	log.Println("[DEBUG] Bailing out because user didn't confirm")
	if noCallback != nil {
		out, err := noCallback()
		return out, shouldContinue, err
	}

	fmt.Println(NegativeAnswer)
	return nil, shouldContinue, nil
}
