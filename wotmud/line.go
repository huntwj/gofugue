package wotmud

import (
	"github.com/huntwj/gofugue/wotmud/prompt"
)

type Line struct {
	Raw        string
	PromptInfo *prompt.Info
	PromptEnd  int
}

// Prompt - get the prompt info for the line
func (l *Line) Prompt() *prompt.Info {
	return nil
}
