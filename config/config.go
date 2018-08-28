package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkstorm/audit_engine/tool"
)

func Version() {
	CFG.Name = "Rule Engine"
	CFG.Version = "version 0.0.1 beta"
	tool.PrettyPrint(CFG.Name, CFG.Version)
}

var CFG struct {
	Test       bool
	ConfigFile string
	TestModule string
	Name       string
	Version    string
	MqConfig   RabbitMqConfig
	DbConfig   MysqlConfig
}

func init() {
	Version()

	//get config file
	pflag.BoolVarP(&CFG.Test, "test", "t", false, "test message publish, need with -m")
	pflag.StringVarP(&CFG.ConfigFile, "config_file", "c", "./config.json", "rule engine config file")
	pflag.StringVarP(&CFG.TestModule, "test_module", "m", "post_audit_msg", "post audit msg test")
	pflag.Parse()

	//read config file
	viper.SetConfigFile(CFG.ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		tool.ErrorPanic(err, "viper read config error")
	}

	//init rabbitmq config
	CFG.MqConfig = RabbitMqConfig{
		Host: viper.GetString("rabbitmq.host"),
		Port: viper.GetInt("rabbitmq.port"),
		User: viper.GetString("rabbitmq.user"),
		Pass: viper.GetString("rabbitmq.pass"),
	}

	//init mysql config
	CFG.DbConfig = MysqlConfig{
		Host: viper.GetString("mysql.host"),
		Port: viper.GetInt("mysql.port"),
		User: viper.GetString("mysql.user"),
		Pass: viper.GetString("mysql.pass"),
	}

	tool.PrettyPrint("config_file", CFG.ConfigFile)
}
