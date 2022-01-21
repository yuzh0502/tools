package main

import (
	"testing"
)

func TestGetRandNum(t *testing.T) {
	for i := 0; i < 10000; i++ {
		n := getRandNum(9)
		if n <= 0 || n > 9 {
			t.Errorf("Error: get rand num: %d", n)
		}
	}
}
