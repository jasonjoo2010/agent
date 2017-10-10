package engine

import (
	"net/http"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stringid"
	"github.com/stretchr/testify/assert"
)

func TestCheckSingleContainerHealthy(t *testing.T) {
	go http.ListenAndServe(":10236", http.NotFoundHandler())
	time.Sleep(100 * time.Millisecond)
	go http.ListenAndServe(":10237", http.NotFoundHandler())
	time.Sleep(100 * time.Millisecond)
	container := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID: stringid.GenerateRandomID(),
		},
		NetworkSettings: &types.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"x": &network.EndpointSettings{
					IPAddress: "127.0.0.1",
				},
			},
		},
	}
	ports := []string{"10236/tcp", "10237/http"}
	state := checkSingleContainerHealthy(container, ports, "/", 404, 3*time.Second)
	assert.True(t, state)
}

func TestCheckAllContainers(t *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	e := mockNewEngine()
	e.checkAllContainers()

	time.Sleep(1 * time.Second)
}

func TestCheckMethodTCP(t *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	assert.False(t, checkTCP(stringid.GenerateRandomID(), []string{"192.168.233.233:10234"}, 2*time.Second))
	go http.ListenAndServe(":10235", http.NotFoundHandler())
	time.Sleep(100 * time.Millisecond)
	assert.True(t, checkTCP(stringid.GenerateRandomID(), []string{"127.0.0.1:10235"}, 2*time.Second))
}

func TestCheckMethodHTTP(t *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// server
	go http.ListenAndServe(":10234", http.NotFoundHandler())
	time.Sleep(100 * time.Millisecond)
	assert.True(t, checkHTTP(stringid.GenerateRandomID(), []string{"http://127.0.0.1:10234/"}, 404, 5*time.Second))
}