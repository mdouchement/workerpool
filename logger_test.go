package workerpool_test

import (
	"testing"

	"github.com/PredicSis/workerpool"
)

func TestLogger(t *testing.T) {
	var l interface{} = &nullLogger{}

	if _, ok := l.(workerpool.Logger); !ok {
		t.Errorf("nullLogger does not implement Logger interface")
	}
}
