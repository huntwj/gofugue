package tokenizer_test

import "testing"
import "github.com/huntwj/gofugue/tflang/tokenizer"

func asSlice(tokens chan tokenizer.Token) []tokenizer.Token {
	tokenSlice := make([]tokenizer.Token, 10)[:0]
	for token := range tokens {
		tokenSlice = append(tokenSlice, token)
	}

	return tokenSlice
}

func assertEqualTokens(t *testing.T, expected, observed tokenizer.Token, message string) {
	if expected.Type != observed.Type {
		t.Errorf("Expected (%v == %v), but found Type mismatch (%v != %v)", expected, observed, expected.Type, observed.Type)
	}
	if expected.Text != observed.Text {
		t.Errorf("Expected (%v == %v), but found Text mismatch (%v != %v)", expected, observed, expected.Type, observed.Type)
	}
}

func assertEqualTokenArrays(t *testing.T, expectedArr, observedArr []tokenizer.Token, message string) {
	for idx, expected := range expectedArr {
		if idx >= len(observedArr) {
			t.Errorf("Expected %d tokens but only received %d", len(expectedArr), len(observedArr))
			return
		}

		observed := observedArr[idx]
		assertEqualTokens(t, expected, observed, message)
	}
	if len(observedArr) > len(expectedArr) {
		t.Errorf("Expected only %d tokens but received %d", len(expectedArr), len(observedArr))
	}
}

func TestTokenizeSimpleCommand(t *testing.T) {
	testStr := "/def"
	expectedTokens := []tokenizer.Token{
		tokenizer.NewToken(tokenizer.SlashCmd, "/def"),
	}

	tokenizer := tokenizer.Tokenize(testStr)
	tokens := asSlice(tokenizer.Tokens())

	assertEqualTokenArrays(t, expectedTokens, tokens, testStr)
}
