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

// A Combatant is someone or something the prompt tells us is in combat with
// or against us. It consists of a Name, which may include spaces and other
// 'weird' characters, and a Health level, which is a standard one word
// indicator of his/her/its health.
type Combatant struct {
	Name   string
	Health string
}

// Combat keeps track of the combatants as shown by the prompt. When in combat,
// the prompt will always show at least the player's and his/her target's
// Combatant info. If the target is attacking someone other than the player
// the Tank Combatant will indicate who or what is tanking for you.
type Combat struct {
	Target Combatant
	Tank   *Combatant
}

// The Info structure keeps track of all the standard information in the
// prompt. Pieces of information that may or may not be present are handled
// as pointers and are nil when the prompt does not contain that information.
// For some information, such as IsRiding, the lack of that information
// actually indicates a falsey value, so a pointer is not used in those cases.
type Info struct {
	IsLit    bool
	IsRiding bool
	Health   string
	Spell    *string
	Moves    string
	Combat   *Combat
}

type line struct {
	raw        string
	promptInfo *Info
	promptEnd  int
	matches    []int
}

func (l line) ss(idx int) string {
	if l.matches[idx] == -1 {
		return ""
	}
	return l.raw[l.matches[idx]:l.matches[idx+1]]
}

// Parse finds a Prompt in a string and returns its data. If no prompt can be
// found nil is returned.
func Parse(str string) (*Info, int) {
	matches := promptRegex.FindStringSubmatchIndex(str)

	if matches == nil {
		matches := weakCandidate.FindStringIndex(str)
		if matches != nil && ignoreRegex.FindStringIndex(str) == nil {
			fmt.Printf("%s failed prompt...\n", str)
			fmt.Println("But weak candidate passed...")
		}

		return nil, 0
	}

	l := line{
		raw:        str,
		promptInfo: &Info{},
		promptEnd:  matches[1],
		matches:    matches,
	}

	if l.matches[priIsLit] != -1 {
		lit := l.ss(priIsLit)
		l.promptInfo.IsLit = (lit == "*")
	}

	if riding := l.ss(priIsRiding); riding != "" {
		l.promptInfo.IsRiding = true
	}

	l.promptInfo.Health = l.ss(priMyHealth)

	if spell := l.ss(priMySpell); spell != "" {
		l.promptInfo.Spell = &spell
	}

	l.promptInfo.Moves = l.ss(priMyMoves)

	if targetName := l.ss(priTargetName); targetName != "" {
		l.promptInfo.Combat = &Combat{}
		l.promptInfo.Combat.Target.Name = targetName
		l.promptInfo.Combat.Target.Health = l.ss(priTargetHealth)
	}

	if tankName := l.ss(priTankName); tankName != "" {
		l.promptInfo.Combat.Tank = &Combatant{}
		l.promptInfo.Combat.Tank.Name = tankName
		l.promptInfo.Combat.Tank.Health = l.ss(priTankHealth)
	}

	return l.promptInfo, l.promptEnd
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
