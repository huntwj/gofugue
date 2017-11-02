package prompt_test

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/huntwj/gofugue/wotmud/prompt"
)

func TestSimpleString(t *testing.T) {
	t.Parallel()

	raw := "This is a simple line"
	line := prompt.Parse(raw)
	if line.Raw != raw {
		t.Error("Raw should always equal input text.")
	}
	if observed := line.PromptEnd; observed != 0 {
		t.Errorf("Simple text line should not have prompt. Expected 0 but got %d\n", observed)
	}
	if line.Prompt != nil {
		t.Errorf("Simple string should not have prompt data")
	}
	if line.PromptEnd != 0 {
		t.Errorf("Simple string should have 0 PromptEnd")
	}
}

func testIsPromptLine(t *testing.T, promptLine string) prompt.Line {
	t.Helper()

	rest := "yet more text on the line"
	raw := promptLine + rest
	line := prompt.Parse(raw)
	if line.Raw != raw {
		t.Error("Raw should always equal input text.")
	}
	promptLen := len(promptLine)
	if observed := line.PromptEnd; observed != promptLen {
		t.Errorf("Line: %s\n", raw)
		t.Errorf("Prompt regex: %s\n", prompt.PromptRegex())
		t.Errorf("Expected '%d' for prompt but got '%d'.", promptLen, observed)
	}
	if line.Prompt == nil {
		t.Errorf("Line: %s\n", raw)
		t.Errorf("Prompt regex: %s\n", prompt.PromptRegex())
		t.Error("Expected non-nil prompt data.\n")
	}

	return line
}

func assertPromptLit(t *testing.T, line prompt.Line, isLit bool) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected line should be lit or not, but did not find prompt!")
	} else if line.Prompt.IsLit != isLit {
		t.Errorf("Room should be lit? %v\n", line.Raw)
	}
}

func assertPromptRiding(t *testing.T, line prompt.Line, expected bool) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected line should be riding or not, but did not find prompt!")
	} else if line.Prompt.IsRiding != expected {
		t.Errorf("Riding mismatch. Expected '%t' observed '%t'\n", expected, !expected)
	}
}

func assertPromptHealth(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check player health, but did not find prompt!")
	} else {
		observed := line.Prompt.Health
		if observed != expected {
			t.Errorf("Health mismatch. Expected '%s' but found '%s'", expected, observed)
		}
	}
}

func assertPromptSpell(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check player spell power, but did not find prompt!")
	} else {
		observed := line.Prompt.Spell
		if observed == nil {
			t.Errorf("Expected spell power but found nil")
		} else if *observed != expected {
			t.Errorf("Spell power mismatch. Expected '%s' but found '%s'", expected, *observed)
		}
	}
}

func assertPromptMoves(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check player moves, but did not find prompt!")
	} else {
		observed := line.Prompt.Moves
		if observed != expected {
			t.Errorf("Moves mismatch. Expected '%s' but found '%s'", expected, observed)
		}
	}
}

func assertPromptCombat(t *testing.T, line prompt.Line, inCombat bool) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check combat status, but did not find prompt!")
	} else if (line.Prompt.Combat == nil) == inCombat {
		t.Errorf("Combat mismatch. Expected %t but observed %t.\n", inCombat, !inCombat)
	}
}

func assertPromptTargetName(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check target name, but did not find prompt!")
	} else if line.Prompt.Combat == nil {
		t.Errorf("Expecting target name but no combat data found")
	} else {
		observed := line.Prompt.Combat.Target.Name
		if observed != expected {
			t.Errorf("Expected target name '%s' but observed '%s'", expected, observed)
		}
	}
}

func assertPromptTargetHealth(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check target health, but did not find prompt!")
	} else if line.Prompt.Combat == nil {
		t.Errorf("Expecting target health but no combat data found")
	} else {
		observed := line.Prompt.Combat.Target.Health
		if observed != expected {
			t.Errorf("Expected target health '%s' but observed '%s'", expected, observed)
		}
	}
}

func assertPromptTank(t *testing.T, line prompt.Line, expected bool) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check tank exists, but did not find prompt!")
	} else if line.Prompt.Combat == nil {
		t.Errorf("Expecting tanked combat but no combat data found")
	} else {
		observed := line.Prompt.Combat.Tank != nil
		if observed != expected {
			t.Errorf("Expected combat with tank '%t' but observed '%t'", expected, observed)
		}
	}
}

func assertPromptTankName(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check tank name, but did not find prompt!")
	} else if line.Prompt.Combat == nil {
		t.Errorf("Expecting tanked combat but no combat data found")
	} else {
		if line.Prompt.Combat.Tank == nil {
			t.Errorf("Expecting tanked combat but no tank data found")
		}

		observed := line.Prompt.Combat.Tank.Name
		if observed != expected {
			t.Errorf("Expected tank name '%s' but observed '%s'", expected, observed)
		}
	}
}

func assertPromptTankHealth(t *testing.T, line prompt.Line, expected string) {
	t.Helper()

	if line.Prompt == nil {
		t.Errorf("Expected to check tank health, but did not find prompt!")
	} else if line.Prompt.Combat == nil {
		t.Errorf("Expecting tanked combat but no combat data found")
	} else {
		if line.Prompt.Combat.Tank == nil {
			t.Errorf("Expecting tanked combat but no tank data found")
		}

		observed := line.Prompt.Combat.Tank.Health
		if observed != expected {
			t.Errorf("Expected tank health '%s' but observed '%s'", expected, observed)
		}
	}
}

func TestPromptString(t *testing.T) {
	t.Parallel()

	prompt := "* HP:Healthy MV:Strong > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, true)
	assertPromptRiding(t, line, false)
	assertPromptHealth(t, line, "Healthy")
	assertPromptMoves(t, line, "Strong")
	assertPromptCombat(t, line, false)
}

func TestPromptStringNoLight(t *testing.T) {
	t.Parallel()

	prompt := "o HP:Healthy MV:Strong > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, false)
	assertPromptRiding(t, line, false)
	assertPromptHealth(t, line, "Healthy")
	assertPromptMoves(t, line, "Strong")
	assertPromptCombat(t, line, false)
}

func TestRidingPrompt(t *testing.T) {
	t.Parallel()

	prompt := "o R HP:Healthy MV:Strong > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, false)
	assertPromptRiding(t, line, true)
	assertPromptHealth(t, line, "Healthy")
	assertPromptMoves(t, line, "Strong")
	assertPromptCombat(t, line, false)
}

func TestChannlerPrompt(t *testing.T) {
	t.Parallel()

	prompt := "* HP:Healthy SP:Bursting MV:Full > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, true)
	assertPromptRiding(t, line, false)
	assertPromptHealth(t, line, "Healthy")
	assertPromptSpell(t, line, "Bursting")
	assertPromptMoves(t, line, "Full")
	assertPromptCombat(t, line, false)
}
func TestCombatWithoutTank(t *testing.T) {
	t.Parallel()

	prompt := "* R HP:Healthy MV:Full - the ancient tree: Critical > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, true)
	assertPromptRiding(t, line, true)
	assertPromptHealth(t, line, "Healthy")
	assertPromptMoves(t, line, "Full")
	assertPromptCombat(t, line, true)
	assertPromptTargetName(t, line, "the ancient tree")
	assertPromptTargetHealth(t, line, "Critical")
	assertPromptTank(t, line, false)
}

func TestCombatWithTank(t *testing.T) {
	t.Parallel()

	prompt := "* R HP:Healthy MV:Fresh - Dal: Scratched - a wild dog: Beaten > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, true)
	assertPromptRiding(t, line, true)
	assertPromptHealth(t, line, "Healthy")
	assertPromptMoves(t, line, "Fresh")
	assertPromptCombat(t, line, true)
	assertPromptTargetName(t, line, "a wild dog")
	assertPromptTargetHealth(t, line, "Beaten")
	assertPromptTank(t, line, true)
	assertPromptTankName(t, line, "Dal")
	assertPromptTankHealth(t, line, "Scratched")
}

func TestComplexCombat(t *testing.T) {
	t.Parallel()

	prompt := "o R HP:Scratched MV:Full - Dal: Healthy - a grayish-green moss: Critical > "
	line := testIsPromptLine(t, prompt)

	assertPromptLit(t, line, false)
	assertPromptRiding(t, line, true)
	assertPromptHealth(t, line, "Scratched")
	assertPromptMoves(t, line, "Full")
	assertPromptCombat(t, line, true)
	assertPromptTargetName(t, line, "a grayish-green moss")
	assertPromptTargetHealth(t, line, "Critical")
	assertPromptTank(t, line, true)
	assertPromptTankName(t, line, "Dal")
	assertPromptTankHealth(t, line, "Healthy")
}

func TestPromptAllHealths(t *testing.T) {
	t.Parallel()

	allHealths := []string{
		"Healthy",
		"Scratched",
		"Hurt",
		"Wounded",
		"Battered",
		"Beaten",
		"Critical",
		"Incapacitated",
		"Dead?",
	}

	for _, health := range allHealths {
		prompt := "o HP:" + health + " MV:Strong > "
		line := testIsPromptLine(t, prompt)

		assertPromptLit(t, line, false)
		assertPromptRiding(t, line, false)
		assertPromptHealth(t, line, health)
		assertPromptMoves(t, line, "Strong")
		assertPromptCombat(t, line, false)
	}
}

func TestPromptAllSpells(t *testing.T) {
	t.Parallel()

	allSpellPowers := []string{
		"Bursting",
		"Full",
		"Strong",
		"Good",
		"Fading",
		"Trickling",
	}

	for _, spellPower := range allSpellPowers {
		prompt := "o HP:Healthy SP:" + spellPower + " MV:Strong > "
		line := testIsPromptLine(t, prompt)

		assertPromptLit(t, line, false)
		assertPromptRiding(t, line, false)
		assertPromptHealth(t, line, "Healthy")
		assertPromptSpell(t, line, spellPower)
		assertPromptMoves(t, line, "Strong")
		assertPromptCombat(t, line, false)
	}
}

func TestPromptAllMovements(t *testing.T) {
	t.Parallel()

	allMovements := []string{
		"Full",
		"Strong",
		"Winded",
		"Tiring",
		"Weary",
		"Haggard",
	}

	for _, movement := range allMovements {
		prompt := "o HP:Healthy MV:" + movement + " > "
		line := testIsPromptLine(t, prompt)

		assertPromptLit(t, line, false)
		assertPromptRiding(t, line, false)
		assertPromptHealth(t, line, "Healthy")
		assertPromptMoves(t, line, movement)
		assertPromptCombat(t, line, false)
	}
}

func testOnLogFile(t *testing.T, fileName string) {

	// fmt.Printf("Opening file: %s\n", fileName)
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		log.Fatal("Could not open file.")
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal("Could not open gzip stream")
	}
	defer gr.Close()

	r := bufio.NewReader(gr)
	// if err != nil {
	// 	log.Fatal("Could not open standard reader")
	// }
	lines := 0
	for str, err := r.ReadString('\n'); ; str, err = r.ReadString('\n') {
		if err == nil {
			prompt.Parse(str)
			lines++
		} else {
			if err != io.EOF {
				t.Errorf("Err: %v", err)
			}
			break
		}
	}
}

func TestOnLogFiles(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping log file tests when short.")
	}
	logDir := "../testdata"
	dir, err := ioutil.ReadDir(logDir)
	if err != nil {
		t.Errorf("Error opening directory: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, fileInfo := range dir {
		fileName := logDir + "/" + fileInfo.Name()
		if !strings.HasSuffix(fileName, ".gz") {
			continue
		}
		wg.Add(1)
		go func() {
			testOnLogFile(t, fileName)
			wg.Done()
		}()
	}
	wg.Wait()
}
