package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	IsDebug       bool `env:"IS_DEBUG" env-default:"false"` // режим отладки
	IsDevelopment bool `env:"IS_DEV" env-default:"false"`   // режим разработчика
	Listen        struct {
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

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Print("gather config")

		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			helpText := "The Art of Development - Monolith Notes System"
			help, _ := cleanenv.GetDescription(instance, &helpText)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return instance
}
