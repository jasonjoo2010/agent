package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectPool(t *testing.T) {
	dropCount := 0
	pool := NewObjectPool(3, func() interface{} {
		return 888
	}, func(obj interface{}) {
		dropCount++
	})

	pool.Put(1)
	pool.Put(2)
	pool.Put(3)
	pool.Put(4) // discard

	assert.Equal(t, 1, dropCount)

	assert.Equal(t, 1, pool.Get())
	assert.Equal(t, 2, pool.Get())
	assert.Equal(t, 3, pool.Get())
	assert.Equal(t, 888, pool.Get())
}
