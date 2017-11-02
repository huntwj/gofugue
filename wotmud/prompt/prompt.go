package prompt

import (
	"fmt"
	"regexp"
)

var weakCandidate = regexp.MustCompile(">")
var ignoreRegex = regexp.MustCompile(`(<Sent: .* >|^<|^ >|Press <Return> to continue|==>|^(\^\[33m)?\w+ (chats|narrates|says|tells you))`)

type substring struct {
	start, end int
}

type Combatant struct {
	Name   string
	Health string
}

type Combat struct {
	Target Combatant
	Tank   *Combatant
}

type PromptData struct {
	IsLit    bool
	IsRiding bool
	Health   string
	Spell    *string
	Moves    string
	Combat   *Combat
}

type Line struct {
	Raw       string
	PromptEnd int
	Prompt    *PromptData
	matches   []int
}

func (line *Line) ss(idx int) string {
	if line.matches[idx] == -1 {
		return ""
	}
	return line.Raw[line.matches[idx]:line.matches[idx+1]]
}

func Parse(str string) Line {
	line := Line{
		Raw:    str,
		Prompt: nil,
	}
	line.matches = promptRegex.FindStringSubmatchIndex(str)
	if line.matches != nil {
		line.Prompt = &PromptData{}
		line.PromptEnd = line.matches[1]

		if line.matches[priIsLit] != -1 {
			lit := line.ss(priIsLit)
			line.Prompt.IsLit = (lit == "*")
		}

		if riding := line.ss(priIsRiding); riding != "" {
			line.Prompt.IsRiding = true
		}

		line.Prompt.Health = line.ss(priMyHealth)

		if spell := line.ss(priMySpell); spell != "" {
			line.Prompt.Spell = &spell
		}

		line.Prompt.Moves = line.ss(priMyMoves)

		if targetName := line.ss(priTargetName); targetName != "" {
			line.Prompt.Combat = &Combat{}
			line.Prompt.Combat.Target.Name = targetName
			line.Prompt.Combat.Target.Health = line.ss(priTargetHealth)
		}

		if tankName := line.ss(priTankName); tankName != "" {
			line.Prompt.Combat.Tank = &Combatant{}
			line.Prompt.Combat.Tank.Name = tankName
			line.Prompt.Combat.Tank.Health = line.ss(priTankHealth)
		}
	} else {
		matches := weakCandidate.FindStringIndex(str)
		if matches != nil && ignoreRegex.FindStringIndex(str) == nil {
			fmt.Printf("%s failed prompt...\n", str)
			fmt.Println("But weak candidate passed...")
		}
	}

	return line
}

var allHealthPattern = `(Healthy|Scratched|Hurt|Wounded|Battered|Beaten|Critical|Incapacitated|Dead\?)`
var allSpellPattern = `(Bursting|Full|Strong|Good|Fading|Trickling)`
var allMovementPattern = `(Full|Fresh|Strong|Winded|Weary|Tiring|Haggard)`
var otherPlayerOrMobPattern = `([\w \-,]+)`
var promptRegex = regexp.MustCompile(
	`^([\*o]) (R )?HP:` + allHealthPattern +
		`(?: SP:` + allSpellPattern +
		`)? MV:` + allMovementPattern +
		`(?:` +
		`(?: - ` + otherPlayerOrMobPattern + `: ` + allHealthPattern + `)?` +
		` - ` + otherPlayerOrMobPattern + `: ` + allHealthPattern +
		`)? > `)

const (
	priIsLit = 2 * (1 + iota)
	priIsRiding
	priMyHealth
	priMySpell
	priMyMoves
	priTankName
	priTankHealth
	priTargetName
	priTargetHealth
)

func PromptRegex() *regexp.Regexp {
	return promptRegex
}
