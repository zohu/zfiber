package zutil

import "testing"

func TestRandomStr(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(RandomStr(4))
	}
	for i := 0; i < 10; i++ {
		t.Log(RandomStr(6))
	}
}
