package config

import (
	"encoding/json"
	"os"
	"time"
)

type Server struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type Services struct {
	UserServiceTimeout time.Duration `json:"user_service_timeout"`
	AuthServiceTimeout time.Duration `json:"auth_service_timeout"`
	BookServiceTimeout time.Duration `json:"book_service_timeout"`
}

type Auth struct {
	Secret          string        `json:"secret,omitempty"`
	SessionLifeTime time.Duration `json:"session_life_time"`
}

type DB struct {
	ConnectionString string        `json:"connection_string"`
	MaxConnections   int           `json:"max_connections"`
	MinConnections   int           `json:"min_connections"`
	MaxIdleTime      time.Duration `json:"max_idle_time"`
	MaxLifetime      time.Duration `json:"max_lifetime"`
}

type Config struct {
	Auth     Auth     `json:"auth"`
	DB       DB       `json:"db"`
	Server   Server   `json:"server"`
	Services Services `json:"services"`
}

var DefaultConfig = Config{
	Server: Server{
		Port:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	},
	Services: Services{
		UserServiceTimeout: 10 * time.Second,
		AuthServiceTimeout: 10 * time.Second,
		BookServiceTimeout: 10 * time.Second,
	},
	Auth: Auth{
		SessionLifeTime: 24 * time.Hour * 30,
		Secret:          "secret",
	},
	DB: DB{
		ConnectionString: "postgres://postgres:postgres@localhost:5432/postgres",
		MaxConnections:   100,
		MinConnections:   10,
		MaxIdleTime:      time.Minute,
		MaxLifetime:      time.Hour,
	},
}

func Parse(filePath string) (config Config, err error) {
	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	err = json.NewDecoder(file).Decode(&config)
	return
}

func MustParse(filePath string) Config {
	config, err := Parse(filePath)
	if err != nil {
		panic(err)
	}
	return config
}
