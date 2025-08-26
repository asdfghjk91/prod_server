package config

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	IsDebug       bool `env:"IS_DEBUG" env-default:"false"` // режим отладки
	IsDevelopment bool `env:"IS_DEV" env-default:"false"`   // режим разработчика
	HTTP          struct {
		IP           string        `yaml:"ip" env:"HTTP-IP"`
		Port         int           `yaml:"ip" env:"HTTP-PORT"`
		ReadTimeout  time.Duration `yaml:"ip" env:"HTTP-READ-TIMEOUT"`
		WriteTimeout time.Duration `yaml:"ip" env:"HTTP-WRITE-TIMEOUT"`
		CORS         struct {
			AllowedMethods     []string `yaml: "allowed_methods" env:"HTTP-CORS-ALLOWEDMETHODS"`
			AllowedOrigins     []string `yaml: "allowed_origins" env:"HTTP-CORS-ALLOWEDORIGINS"`
			AllowCredentials   bool     `yaml: "allowed_credentials" env:"HTTP-CORS-ALLOWCREDENTIALS"`
			AllowedHeaders     []string `yaml: "allowed_headers" env:"HTTP-CORS-ALLOWEDHEADERS"`
			OptionsPassthrough bool     `yaml: "options_passthrough" env:"HTTP-CORS-OPTIONSPASSTHROUGH"`
			ExposedHeaders     []string `yaml: "exposed_headers" env:"HTTP-CORS-EXPOSEDHEADERS"`
			Debug              bool     `yaml: "debug" env:"HTTP-CORS-DEBUG"`
		} `yaml: "cors"`
	}

	Listen struct {
		Type       string `env:"LISTEN_TYPE" env-default:"port" env-description:"port or sock. If sock then env SOCKET_FILE env is required"` // как слушать port или socket
		BindIP     string `env:"BIND_IP" env-default:"0.0.0.0"`                                                                               // IP сервера (0.0.0.0 на все интерфейсы)
		Port       string `env:"PORT" env-default:9090`                                                                                       // порт для запуска
		SocketFile string `env:"SOCKET_FILE" env-default:"app.sock"`
	}
	AppConfig struct {
		LogLevel  string `env:"LOG_LEVEL" env-deault:trace"`
		AdminUser struct {
			Email    string `env:"ADMIN_EMAIL" env-default:"admin"`
			Password string `env:"ADMIN_PWD" env-default:"admin"`
		}
	}

	PostgresqSQL struct {
		Username string `env:"PSQL_USERNAME" env-required:"true"`
		Password string `env:"PSQL_PASSWORD" env-required:"true"`
		Host     string `env:"PSQL_HOST" env-required:"true"`
		Port     string `env:"PSQL_PORT" env-required:"true"`
		Database string `env:"PSQL_DATABASE" env-required:"true"`
	}
}

const (
	EnvConfigPathName  = "CONFIG-PATH"
	FlagConfigPathName = "config"
)

var configPath string
var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		flag.StringVar(&configPath, FlagConfigPathName, "configs/config.local.yaml", "this app config file")
		flag.Parse()

		log.Print("config init")

		if configPath == "" {
			configPath = os.Getenv(EnvConfigPathName)
		}

		if configPath == "" {
			log.Fatalf("config path is required")
		}
		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			helpText := "The Art of Development - Production Service"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
