package pinner_test

import (
	"testing"

	"github.com/StabbyCutyou/a_conversation_with_dave"
)

func TestPinner(t *testing.T) {
	pinner.Register("github.com/StabbyCutyou/buffstreams", "= 1.0")
	pinner.Register("github.com/constabulary/gb", "= 0.1.0")

	err := pinner.Pin()
	if err != nil {
		t.Error(err)
	}

}
