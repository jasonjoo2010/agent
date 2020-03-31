package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormal(t *testing.T) {
	addr_list := []string{"1", "2", "3", "4"}
	backends := NewHashBackends(addr_list)
	assert.Equal(t, uint32(len(addr_list)), backends.Len())
	addr := backends.Get("test", 0)
	assert.NotEmpty(t, addr)
	assert.NotEqual(t, addr, backends.Get("test", 1))
}

func TestEmpty(t *testing.T) {
	backends := NewHashBackends([]string{})
	assert.Equal(t, uint32(0), backends.Len())
	assert.Empty(t, backends.Get("test", 0))
}
