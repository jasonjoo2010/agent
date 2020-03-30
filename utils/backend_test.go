package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBackend(t *testing.T) {
	b, err := ParseBackend("")
	assert.Nil(t, b)
	assert.Equal(t, EmptyBackend, err)

	b, err = ParseBackend("udp://test")
	assert.Nil(t, b)
	assert.Equal(t, IllegalPort, err)

	b, err = ParseBackend("tcp://:333")
	assert.Nil(t, b)
	assert.Equal(t, IllegalFormat, err)

	b, err = ParseBackend("tcp://test:")
	assert.Nil(t, b)
	assert.Equal(t, IllegalPort, err)

	b, err = ParseBackend("http://test:test")
	assert.Nil(t, b)
	assert.Equal(t, IllegalFormat, err)

	b, err = ParseBackend("http://test:80")
	assert.Nil(t, b)
	assert.Equal(t, IllegalType, err)

	// normal, tcp
	b, err = ParseBackend("tcp://test:803")
	assert.NotNil(t, b)
	assert.Nil(t, err)
	assert.Equal(t, TCP, b.Type)
	assert.Equal(t, "test", b.Host)
	assert.Equal(t, 803, b.Port)

	// normal, udp
	b, err = ParseBackend("udp://192.168.3.1:65535")
	assert.NotNil(t, b)
	assert.Nil(t, err)
	assert.Equal(t, UDP, b.Type)
	assert.Equal(t, "192.168.3.1", b.Host)
	assert.Equal(t, 65535, b.Port)

	// normal, journal
	b, err = ParseBackend("journal://www.demo-site.com:321")
	assert.NotNil(t, b)
	assert.Nil(t, err)
	assert.Equal(t, Journal, b.Type)
	assert.Equal(t, "www.demo-site.com", b.Host)
	assert.Equal(t, 321, b.Port)
	assert.Equal(t, "www.demo-site.com:321", b.HostWithPort)

	// normal, blackhole, could be "blackhole://"
	b, err = ParseBackend("blackhole://www.demo-site.com:321")
	assert.NotNil(t, b)
	assert.Nil(t, err)
	assert.Equal(t, BlackHole, b.Type)
	assert.Equal(t, "", b.Host)
	assert.Equal(t, 0, b.Port)
}

func TestParseBackends(t *testing.T) {
	// normal
	backends, err := ParseBackends([]string{"tcp://127.0.0.1:333", "udp://192.168.0.1:3333"})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(backends))

	backends, err = ParseBackends([]string{})
	assert.Equal(t, EmptyBackend, err)
	assert.Nil(t, backends)
}
