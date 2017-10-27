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
}

func testIsPromptLine(t *testing.T, promptLine string) {
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
}
func TestPromptString(t *testing.T) {
	t.Parallel()

	prompt := "* HP:Healthy MV:Strong > "
	testIsPromptLine(t, prompt)
}

func TestPromptStringNoLight(t *testing.T) {
	t.Parallel()

	prompt := "o HP:Healthy MV:Strong > "
	testIsPromptLine(t, prompt)
}

func TestRidingPrompt(t *testing.T) {
	t.Parallel()

	prompt := "o R HP:Healthy MV:Strong > "
	testIsPromptLine(t, prompt)
}

func TestCombatWithoutTank(t *testing.T) {
	t.Parallel()

	prompt := "* R HP:Healthy MV:Full - the ancient tree: Critical > "
	testIsPromptLine(t, prompt)
}

func TestCombatWithTank(t *testing.T) {
	t.Parallel()

	prompt := "* R HP:Healthy MV:Fresh - Dal: Scratched - a wild dog: Beaten > "
	testIsPromptLine(t, prompt)
}

func TestComplexCombat(t *testing.T) {
	t.Parallel()

	prompt := "* R HP:Scratched MV:Full - Dal: Healthy - a grayish-green moss: Critical > "
	testIsPromptLine(t, prompt)
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
		testIsPromptLine(t, prompt)
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
		testIsPromptLine(t, prompt)
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
