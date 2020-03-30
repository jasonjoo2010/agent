package utils

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Type int

const (
	TCP       Type = 1
	UDP       Type = 2
	Journal   Type = 3
	BlackHole Type = 4
)

var (
	EmptyBackend  error = errors.New("Backend can not be empty")
	IllegalFormat error = errors.New("Illegal backend format, it must be in type://host:port form")
	IllegalPort   error = errors.New("Port should be a positive number")
	IllegalType   error = errors.New("Backend type specified is not supported")
)

type Backend struct {
	Type         Type
	Host         string
	Port         int
	HostWithPort string
}

// NewBackend is just a convenient way to create a Backend and will not do any validation.
func NewBackend(t Type, host string, port int) *Backend {
	return &Backend{
		Type:         t,
		Host:         host,
		Port:         port,
		HostWithPort: fmt.Sprintf("%s:%d", host, port),
	}
}

// ParseBackend parses backend into Backend
func ParseBackend(backend string) (*Backend, error) {
	if backend == "" {
		return nil, EmptyBackend
	}

	var (
		t    Type
		host string
		port int
	)

	u, err := url.Parse(backend)
	if err != nil || len(u.Hostname()) < 1 {
		return nil, IllegalFormat
	}

	switch strings.ToLower(u.Scheme) {
	case "udp":
		t = UDP
	case "tcp":
		t = TCP
	case "blackhole":
		// ignore host and port
		t = BlackHole
		return NewBackend(t, "", 0), nil
	case "journal":
		t = Journal
	default:
		return nil, IllegalType
	}

	host = u.Hostname()

	port, err = strconv.Atoi(u.Port())
	if err != nil || port < 1 {
		return nil, IllegalPort
	}

	return NewBackend(t, host, port), nil
}

// ParseBackend parses []string into []*Backends
func ParseBackends(backends []string) ([]*Backend, error) {
	if backends == nil || len(backends) == 0 {
		return nil, EmptyBackend
	}
	list := make([]*Backend, len(backends))
	for i := 0; i < len(backends); i++ {
		b, err := ParseBackend(backends[i])
		if err != nil {
			return nil, err
		}
		list[i] = b
	}
	return list, nil
}
