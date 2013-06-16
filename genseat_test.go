package main
import (
	"testing"
)
func TestGetProcInputs(t *testing.T) {
	devs := getProcInputs()
	if len(devs)==0 {
		t.Error("get zero devs")
	}
}
