package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

var Engine struct {
	Name     string
	Version  string
	MqConfig RabbitMqConfig
	DbConfig MysqlConfig
}

func Version() {
	Engine.Name = "Obs Rule Engine"
	Engine.Version = "ver 0.0.1 beta"
	log.Printf("%s %s\n", Engine.Name, Engine.Version)
}

func init() {
	Version()

	//get config file
	cf := pflag.StringP("config_file", "c", "./config.json", "rule engine config file")
	pflag.Parse()

	//read config file
	viper.SetConfigFile(*cf)
	if err := viper.ReadInConfig(); err != nil {
		tool.ErrorLog(err, "viper read config error")
	}

	//init rabbitmq config
	Engine.MqConfig = RabbitMqConfig{
		Host: viper.GetString("rabbitmq.host"),
		Port: viper.GetInt("rabbitmq.port"),
		User: viper.GetString("rabbitmq.user"),
		Pass: viper.GetString("rabbitmq.pass"),
	}

	//init mysql config
	Engine.DbConfig = MysqlConfig{
		Host: viper.GetString("mysql.host"),
		Port: viper.GetInt("mysql.port"),
		User: viper.GetString("mysql.user"),
		Pass: viper.GetString("mysql.pass"),
	}

	tool.PrettyPrint("config_file", *cf)
}
