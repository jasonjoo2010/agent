package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestConfigEmpty(t *testing.T) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "pidfile",
				Required: true,
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		config := Config{}
		config.PrepareConfig(c)
		// default values
		assert.Equal(t, 3, config.HealthCheckTimeout)
		assert.Equal(t, 10, config.HealthCheckInterval)
		assert.Equal(t, 60, config.HealthCheckCacheTTL)
		assert.NotEmpty(t, config.HostName)
		assert.Equal(t, "a.pid", config.PidFile)
		return nil
	}
	app.Run([]string{
		"cmd",
		"--pidfile",
		"a.pid",
	})
}

func TestConfigStrings(t *testing.T) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "pidfile",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "hostname",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "log-forwards",
				Required: false,
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		config := Config{}
		config.PrepareConfig(c)
		assert.Equal(t, "demo-machine", config.HostName)
		assert.NotEmpty(t, config.HostName)
		assert.Equal(t, 3, len(config.Log.Forwards))
		assert.Contains(t, config.Log.Forwards, "192.168.0.1:444")
		return nil
	}
	app.Run([]string{
		"cmd",
		"--pidfile",
		"a.pid",
		"--hostname",
		"demo-machine",
		"--log-forwards",
		"tcp://192.168.0.1:444",
		"--log-forwards",
		"udp://192.168.0.2:444",
		"--log-forwards",
		"journal://192.168.0.3:444",
	})
}

func TestConfigOthers(t *testing.T) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "pidfile",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "health-check-interval",
				Required: false,
			},
			&cli.IntFlag{
				Name:     "health-check-ttl",
				Required: false,
			},
			&cli.Int64Flag{
				Name:     "metrics-step",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "log-stdout",
				Required: false,
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		config := Config{}
		config.PrepareConfig(c)
		assert.Equal(t, 31, config.HealthCheckInterval)
		assert.Equal(t, 60, config.HealthCheckCacheTTL)
		assert.Equal(t, int64(333), config.Metrics.Step)
		assert.True(t, config.Log.Stdout)
		assert.NotEmpty(t, config.HostName)
		return nil
	}
	app.Run([]string{
		"cmd",
		"--pidfile",
		"a.pid",
		"--health-check-interval",
		"31",
		"--health-check-ttl",
		"31",
		"--metrics-step",
		"333",
		"--log-stdout",
		"yes",
	})
}
