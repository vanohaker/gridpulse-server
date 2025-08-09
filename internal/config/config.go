package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type ConfigYaml struct {
	Redis     Redis    `yaml:"redis"`
	Postgres  Postgres `yaml:"postgres"`
	AppRes    AppRes   `yaml:"app"`
	Jwtsecret string   `yaml:"jwtsecret"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database int    `yaml:"db"`
}

type Postgres struct {
	DbString string
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Dbname   string `yaml:"dbname"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type AppRes struct {
	Bind string `yaml:"bind"`
	Port string `yaml:"port"`
}

func LoadConfig(logger zerolog.Logger, configPath *string) (*ConfigYaml, error) {
	config := &ConfigYaml{}
	file, err := os.Open(*configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	err = viper.ReadConfig(file)
	if err != nil {
		return nil, err
	}
	viper.SetDefault("app.bind", "0.0.0.0")
	viper.SetDefault("app.port", 3000)
	config.Redis.Host = viper.GetString("redis.host")
	config.Redis.Port = viper.GetInt("redis.port")
	config.Redis.Database = viper.GetInt("redis.db")
	config.Postgres.Host = viper.GetString("postgres.host")
	config.Postgres.Port = viper.GetInt("postgres.port")
	config.Postgres.Dbname = viper.GetString("postgres.dbname")
	config.Postgres.User = viper.GetString("postgres.user")
	config.Postgres.Password = viper.GetString("postgres.password")
	config.Postgres.DbString = func() string {
		// postgres://user:password@qwerty.us-east-1.redshift.amazonaws.com:5439/db
		return fmt.Sprintf("postgres://%s:%s@%s:%v/%s", config.Postgres.User, config.Postgres.Password, config.Postgres.Host, config.Postgres.Port, config.Postgres.Dbname)
	}()
	config.AppRes.Bind = viper.GetString("app.bind")
	config.AppRes.Port = viper.GetString("app.port")
	config.Jwtsecret = viper.GetString("jwtsecret")
	return config, nil
}
