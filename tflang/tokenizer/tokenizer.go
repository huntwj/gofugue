package tokenizer

import (
	"io"
	"strings"
)

// Token - Keeps track of all information relating to a lexical token
type Token struct {
	Type rune
	Text string
}

const (
	// SlashCmd - A TinyFugue slash command such as /def
	SlashCmd = -(iota + 1)
)

// Tokenizer - Data structure managing the conversion from source data into
// lexical tokens.
type Tokenizer struct {
	reader io.RuneReader
}

func isValidSlashCommandRune(ch rune) bool {
	return true
}

func (t *Tokenizer) lex(tokenChan chan Token) {
	defer close(tokenChan)

	var readAhead rune
	readAheadSize := 0
	ch, _, err := t.reader.ReadRune()
outerLexLoop:
	for err == nil {
		switch ch {
		case '/':
			tokenBuff := make([]rune, 10)[:1]
			tokenBuff[0] = ch
			readAhead, readAheadSize, err = t.reader.ReadRune()
			for err == nil &&
				readAheadSize > 0 &&
				isValidSlashCommandRune(readAhead) {
				tokenBuff = append(tokenBuff, readAhead)
				readAhead, readAheadSize, err = t.reader.ReadRune()
			}
			if len(tokenBuff) > 1 {
				tokenChan <- Token{
					Type: SlashCmd,
					Text: string(tokenBuff),
				}
			} else {
				tokenChan <- Token{
					Type: '/',
					Text: "/",
				}
			}
			if err != nil {
				break outerLexLoop
			}
		}
		if readAheadSize == 0 {
			ch, _, err = t.reader.ReadRune()
		} else {
			ch, _ = readAhead, readAheadSize
			readAheadSize = 0
		}
	}
}

// Tokens - Start a goroutine that will send Tokens to the returned channel
func (t *Tokenizer) Tokens() chan Token {
	tokenChan := make(chan Token, 10)
	go t.lex(tokenChan)
	return tokenChan
}

// NewToken - Convenience func to construct a new token
func NewToken(tokenType rune, tokenText string) Token {
	return Token{
		Type: tokenType,
		Text: tokenText,
	}
}

// Tokenize - Create a tokenizer for a string input
func Tokenize(sourceStr string) Tokenizer {
	return Tokenizer{
		reader: strings.NewReader(sourceStr),
	}
}
