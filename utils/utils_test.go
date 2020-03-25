package utils

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/projecteru2/agent/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckExistsError(t *testing.T) {
	// not exist return nil
	assert.Nil(t, CheckExistsError(client.Error{
		Code: client.ErrorCodeNodeExist,
	}))

	// other errors keep it
	assert.Equal(t, client.Error{
		Code: client.ErrorCodeEventIndexCleared,
	}, CheckExistsError(client.Error{
		Code: client.ErrorCodeEventIndexCleared,
	}))
	assert.Equal(t, io.EOF, CheckExistsError(io.EOF))
}

func TestMakeDockerClient(t *testing.T) {
	client, err := MakeDockerClient(&types.Config{
		Docker: types.DockerConfig{
			Endpoint: "tcp://127.0.0.1:2379",
		},
	})
	assert.Nil(t, err)
	assert.Contains(t, client.CustomHTTPHeaders()["User-Agent"], "eru-agent-")
}

func TestWritePid(t *testing.T) {
	pidPath, err := ioutil.TempFile(os.TempDir(), "pid-")
	assert.NoError(t, err)

	WritePid(pidPath.Name())

	f, err := os.Open(pidPath.Name())
	assert.NoError(t, err)

	content, err := ioutil.ReadAll(f)
	assert.NoError(t, err)

	pid := strconv.Itoa(os.Getpid())
	assert.Equal(t, pid, string(content))

	os.Remove(pidPath.Name())
}

func TestGetAppInfo(t *testing.T) {
	containerName := "eru-stats_api_EAXPcM"
	name, entrypoint, ident, err := GetAppInfo(containerName)
	assert.NoError(t, err)

	assert.Equal(t, name, "eru-stats")
	assert.Equal(t, entrypoint, "api")
	assert.Equal(t, ident, "EAXPcM")

	containerName = "api_EAXPcM"
	_, _, _, err = GetAppInfo(containerName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid container name")
}

func TestMax(t *testing.T) {
	assert.Equal(t, int64(1), Max(1, 0))
	assert.Equal(t, int64(1), Max(0, 1))
	assert.Equal(t, int64(2), Max(1, 2))
	assert.Equal(t, int64(2), Max(2, 1))
	assert.Equal(t, int64(2), Max(2, 2))
	assert.Equal(t, int64(2), Max(2, -2))
	assert.Equal(t, int64(-1), Max(-1, -2))
	assert.Equal(t, int64(-1), Max(-2, -1))
}
