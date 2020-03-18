package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrtoi(t *testing.T) {
	r := []rune("10")
	n, e := strtoi(&r)
	assert.Nil(t, e)
	assert.Equal(t, 10, n)
	assert.Equal(t, []rune{}, r)
	r = []rune("-5")
	n, e = strtoi(&r)
	assert.Nil(t, e)
	assert.Equal(t, -5, n)
	assert.Equal(t, []rune{}, r)
	r = []rune("-")
	n, e = strtoi(&r)
	assert.NotNil(t, e)
	r = []rune("13+")
	n, e = strtoi(&r)
	assert.Nil(t, e)
	assert.Equal(t, 13, n)
	assert.Equal(t, []rune("+"), r)
	r = []rune("   28 +")
	n, e = strtoi(&r)
	assert.Nil(t, e)
	assert.Equal(t, 28, n)
	assert.Equal(t, []rune(" +"), r)
}
