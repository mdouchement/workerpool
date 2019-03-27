package workerpool_test

import (
	"testing"

	"github.com/mdouchement/workerpool"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var l interface{} = &nullLogger{}

	_, ok := l.(workerpool.Logger)
	assert.True(t, ok, "nullLogger does not implement Logger interface")
}
