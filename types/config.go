package types

import (
	"os"
	"strings"

	coretypes "github.com/projecteru2/core/types"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// DockerConfig contain endpoint
type DockerConfig struct {
	Endpoint string `yaml:"endpoint" required:"true"`
}

// MetricsConfig contain metrics config
type MetricsConfig struct {
	Step      int64    `yaml:"step" required:"true" default:"10"`
	Transfers []string `yaml:"transfers"`
}

// APIConfig contain api config
type APIConfig struct {
	Addr string `yaml:"addr"`
}

// LogConfig contain log config
type LogConfig struct {
	Forwards    []string `yaml:"forwards"`
	Stdout      bool     `yaml:"stdout"` // deprecated
	Connections int      `yaml:"connections" default:"10"`
	Ratelimit   int      `yaml:"ratelimit" default:"-1"`
	BufferSize  int      `yaml:"buffer_size" default:"5000"`
}

// Config contain all configs
type Config struct {
	PidFile             string               `yaml:"pid" required:"true" default:"/tmp/agent.pid"`
	HealthCheckInterval int                  `yaml:"health_check_interval"`
	HealthCheckTimeout  int                  `yaml:"health_check_timeout"`
	HealthCheckCacheTTL int                  `yaml:"health_check_cache_ttl"`
	Core                string               `yaml:"core" required:"true"`
	Auth                coretypes.AuthConfig `yaml:"auth"`
	HostName            string               `yaml:"-"`

	Docker  DockerConfig
	Metrics MetricsConfig
	API     APIConfig
	Log     LogConfig
}

func setString(attr *string, c *cli.Context, key string) {
	val := c.String(key)
	if val != "" {
		*attr = val
	}
}

func setStringSlice(attr *[]string, c *cli.Context, key string) {
	val := c.StringSlice(key)
	if len(val) > 0 {
		*attr = val
	}
}

func setBool(attr *bool, c *cli.Context, key string) {
	val := c.String(key)
	if val != "" {
		*attr = strings.EqualFold(val, "yes") || strings.EqualFold(val, "true")
	}
}

func setInt(attr *int, c *cli.Context, key string) {
	val := c.Int(key)
	if val != 0 {
		*attr = val
	}
}

func setInt64(attr *int64, c *cli.Context, key string) {
	val := c.Int64(key)
	if val != 0 {
		*attr = val
	}
}

//PrepareConfig 从cli覆写并做准备
func (config *Config) PrepareConfig(c *cli.Context) {
	setString(&config.HostName, c, "hostname")
	setString(&config.Core, c, "core-endpoint")
	setString(&config.Auth.Username, c, "core-username")
	setString(&config.Auth.Password, c, "core-password")
	setString(&config.PidFile, c, "pidfile")
	setInt(&config.HealthCheckInterval, c, "health-check-interval")
	setInt(&config.HealthCheckTimeout, c, "health-check-timeout")
	setString(&config.Docker.Endpoint, c, "docker-endpoint")
	setInt64(&config.Metrics.Step, c, "metrics-step")
	setStringSlice(&config.Metrics.Transfers, c, "metrics-transfers")
	setString(&config.API.Addr, c, "api-addr")
	setStringSlice(&config.Log.Forwards, c, "log-forwards")
	//setBool(&config.Log.Stdout, c, "log-stdout")
	setInt(&config.Log.Connections, c, "log-forwards-connections")
	setInt(&config.Log.BufferSize, c, "log-forwards-buffer-size")
	setInt(&config.Log.Ratelimit, c, "log-forwards-ratelimit")

	//validate
	if config.HostName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		} else {
			config.HostName = hostname
		}
	}
	if config.PidFile == "" {
		log.Fatal("need to set pidfile")
	}
	if config.HealthCheckTimeout < 1 {
		config.HealthCheckTimeout = 3
	}
	if config.HealthCheckInterval < 10 {
		config.HealthCheckInterval = 10
	}
	if config.HealthCheckCacheTTL < config.HealthCheckInterval*3/2 {
		config.HealthCheckCacheTTL = 60
	}
}
