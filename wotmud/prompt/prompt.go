package prompt

import (
	"fmt"
	"regexp"
)

var weakCandidate = regexp.MustCompile(">")
var ignoreRegex = regexp.MustCompile("(<Sent: .* >|^<|^ >|Press <Return> to continue)")

type substring struct {
	start, end int
}

type Line struct {
	Raw       string
	PromptEnd int
}

func Parse(str string) Line {
	line := Line{
		Raw: str,
	}
	matches := promptRegex.FindAllStringIndex(str, -1)
	if matches != nil {
		line.PromptEnd = matches[0][1]
	} else {
		matches = weakCandidate.FindAllStringIndex(str, -1)
		if matches != nil && ignoreRegex.FindAllStringIndex(str, -1) == nil {
			fmt.Printf("%s failed prompt...\n", str)
			fmt.Println("But weak candidate passed...")
		}
	}

	return line
}

var allHealthPattern = `(Healthy|Scratched|Hurt|Wounded|Battered|Beaten|Critical|Incapacitated|Dead\?)`
var allMovementPattern = `(Full|Fresh|Strong|Winded|Weary|Tiring|Haggard)`
var otherPlayerOrMobPattern = `([\w \-,]+)`
var promptRegex = regexp.MustCompile(
	`[\*o] (R )?HP:` + allHealthPattern +
		` MV:` + allMovementPattern +
		`(` +
		`( - ` + otherPlayerOrMobPattern + `: ` + allHealthPattern + `)?` +
		` - ` + otherPlayerOrMobPattern + `: ` + allHealthPattern +
		`)? > `)

func PromptRegex() *regexp.Regexp {
	return promptRegex
}
