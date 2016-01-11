package pinner_test

import (
	"testing"

	"github.com/StabbyCutyou/pinner"
)

func TestPinner(t *testing.T) {
	pinner.Register("github.com/StabbyCutyou/lib_a", "~> 0.4")
	pinner.Register("github.com/StabbyCutyou/lib_b", "= 2.1.4")

	err := pinner.Pin()
	if err != nil {
		t.Error(err)
	}

}
