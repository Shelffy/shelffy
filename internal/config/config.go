package config

import (
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v3"
)

const (
	EnvAuthSecret         = "AUTH_SECRET"
	EnvDBConnectionString = "DB_CONNECTION_STRING"
	EnvS3APIEndpoint      = "S3_API_ENDPOINT"
	EnvS3Bucket           = "S3_BUCKET"
	EnvS3AccessKeyID      = "S3_ACCESS_KEY_ID"
	EnvS3SecretKey        = "S3_SECRET_KEY"
	EnvS3AccountID        = "S3_ACCOUNT_ID"
	EnvS3TokenValue       = "S3_TOKEN_VALUE"
	EnvNATSUrl            = "NATS_URL"
)

type Endpoints struct {
	API APIEndpoints `json:"api" yaml:"api"`
}

type APIEndpoints struct {
	Base string             `json:"base" yaml:"base"`
	V1   APIV1BaseEndpoints `json:"v1" yaml:"v1"`
	GQL  GQLEndpoints       `json:"gql" yaml:"gql"`
}

type APIV1BaseEndpoints struct {
	Base  string `json:"base" yaml:"base"`
	Books string `json:"books" yaml:"books"`
	Users string `json:"users" yaml:"users"`
	Auth  string `json:"auth" yaml:"auth"`
}

type GQLEndpoints struct {
	Base  string `json:"base" yaml:"base"`
	Query string `json:"query" yaml:"query"`
}

type Server struct {
	Port         string        `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	Endpoints    Endpoints     `json:"endpoints" yaml:"endpoints"`
}

type Services struct {
	UserServiceTimeout time.Duration `json:"user_service_timeout" yaml:"user_service_timeout"`
	AuthServiceTimeout time.Duration `json:"auth_service_timeout" yaml:"auth_service_timeout"`
	BookServiceTimeout time.Duration `json:"book_service_timeout" yaml:"book_service_timeout"`
}

type Auth struct {
	Secret          string        `json:"secret,omitempty" yaml:"secret,omitempty"`
	SessionLifeTime time.Duration `json:"session_life_time" yaml:"session_life_time"`
}

type DB struct {
	ConnectionString string        `json:"connection_string" yaml:"connection_string"`
	MaxConnections   int           `json:"max_connections" yaml:"max_connections"`
	MinConnections   int           `json:"min_connections" yaml:"min_connections"`
	MaxIdleTime      time.Duration `json:"max_idle_time" yaml:"max_idle_time"`
	MaxLifetime      time.Duration `json:"max_life_time" yaml:"max_lifetime"`
}

type S3 struct {
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`
	AccessSecretKey string `json:"secret_key" yaml:"secret_key"`
	AccountID       string `json:"account_id" yaml:"account_id"`
	APIEndpoint     string `json:"api_endpoint" yaml:"api_endpoint"`
	BooksBucket     string `json:"books_bucket" yaml:"books_bucket"`
	TokenValue      string `json:"token_value" yaml:"token_value"`
}

type NATS struct {
	URL string `json:"nats" yaml:"url"`
}

type Config struct {
	Auth     Auth     `json:"auth" yaml:"auth"`
	DB       DB       `json:"db" yaml:"db"`
	Server   Server   `json:"server" yaml:"server"`
	Services Services `json:"services" yaml:"services"`
	S3       S3       `json:"s3" yaml:"s3"`
	NATS     NATS     `json:"nats" yaml:"nats"`
	Debug    bool     `json:"debug" yaml:"debug"`
}

var DefaultConfig = Config{
	Server: Server{
		Port:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Endpoints: Endpoints{
			API: APIEndpoints{
				Base: "/api",
				V1: APIV1BaseEndpoints{
					Base:  "/v1",
					Books: "/books",
					Users: "/users",
					Auth:  "/auth",
				},
				GQL: GQLEndpoints{
					Base:  "/gql",
					Query: "/query",
				},
			},
		},
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
	S3: S3{
		AccessKeyID:     "",
		AccessSecretKey: "",
		AccountID:       "",
		APIEndpoint:     "",
		BooksBucket:     "",
		TokenValue:      "",
	},
	NATS:  NATS{URL: nats.DefaultURL},
	Debug: true,
}

func setENV(config *Config) {
	if val := os.Getenv(EnvAuthSecret); val != "" {
		config.Auth.Secret = val
	}
	if val := os.Getenv(EnvDBConnectionString); val != "" {
		config.DB.ConnectionString = val
	}
	if val := os.Getenv(EnvS3APIEndpoint); val != "" {
		config.S3.APIEndpoint = val
	}
	if val := os.Getenv(EnvS3Bucket); val != "" {
		config.S3.BooksBucket = val
	}
	if val := os.Getenv(EnvS3AccessKeyID); val != "" {
		config.S3.AccessKeyID = val
	}
	if val := os.Getenv(EnvS3SecretKey); val != "" {
		config.S3.AccessSecretKey = val
	}
	if val := os.Getenv(EnvS3AccountID); val != "" {
		config.S3.AccountID = val
	}
	if val := os.Getenv(EnvS3TokenValue); val != "" {
		config.S3.TokenValue = val
	}
	if val := os.Getenv(EnvNATSUrl); val != "" {
		config.NATS.URL = val
	}
}

// Parse parses config.
// If some variables were provided in environment variables, they will be used in the config
func Parse(filePath string) (config Config, err error) {
	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return
	}
	if err = yaml.NewDecoder(file).Decode(&config); err != nil {
		return
	}
	setENV(&config)
	return
}

func MustParse(filePath string) Config {
	config, err := Parse(filePath)
	if err != nil {
		panic(err)
	}
	return config
}
