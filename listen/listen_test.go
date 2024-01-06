package listen

import (
	"os"
	"testing"
)

func TestConcatWav(t *testing.T) {
	i1, _ := os.ReadFile(".\\test.wav")
	i2, _ := os.ReadFile(".\\test.wav")

	res := ConcatWav(i1, i2)

	os.WriteFile(".\\testR.wav", res, 0644)
}
