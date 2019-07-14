package helpers

import (
	"testing"
)

func TestPow(t *testing.T) {
	if Pow(2, 2) != 4 {
		t.Fatal("POW yielded incorrect answer")
	}
	if Pow(2, 3) != 8 {
		t.Fatal("POW yielded incorrect answer")
	}
	if Pow(2, 4) != 16 {
		t.Fatal("POW yielded incorrect answer")
	}
}
