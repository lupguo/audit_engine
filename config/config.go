package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

type EngineInfo struct {
	Name    string
	Version string
}

type CFG struct {
	cmd        CmdArgs
	EInfo      EngineInfo
	Test       bool
	ConfigFile string
	RabbitMq   map[string]rabbit.Config
	Mysql      mydb.Config
}

//version info
func (cfg *CFG) GetVersion(egi EngineInfo) string {
	return fmt.Sprintf("%s, %s", egi.Name, egi.Version)
}

//config init
func (cfg *CFG) InitByCmd(cmd CmdArgs) {
	//read config file
	viper.SetConfigFile(cmd.Cfg)
	err := viper.ReadInConfig()
	tool.FatalLog(err, "viper read config error")

	//test
	cfg.cmd = cmd
	cfg.Test = cmd.T
	cfg.ConfigFile = cmd.Cfg

	//version
	cfg.EInfo = EngineInfo{
		Name:    viper.GetString("name"),
		Version: viper.GetString("version"),
	}

	//init rabbitmq config
	cfg.RabbitMq = make(map[string]rabbit.Config)
	for _, v := range []string{"soa", "gb"} {
		cfg.RabbitMq[v] = rabbit.Config{
			Host:  viper.GetString("rabbitmq." + v + ".host"),
			Port:  viper.GetInt("rabbitmq." + v + ".port"),
			User:  viper.GetString("rabbitmq." + v + ".user"),
			Pass:  viper.GetString("rabbitmq." + v + ".pass"),
			Vhost: viper.GetString("rabbitmq." + v + ".vhost"),
		}
	}

	//init mysql config
	cfg.Mysql = mydb.Config{
		Host:        viper.GetString("mysql.host"),
		Port:        viper.GetInt("mysql.port"),
		User:        viper.GetString("mysql.user"),
		Pass:        viper.GetString("mysql.pass"),
		DbName:      viper.GetString("mysql.dbname"),
		Protocol:    viper.GetString("mysql.protocol"),
		ConnMaxLife: viper.GetInt("mysql.conn_max_life"),
	}

}

//show all info
func (cfg *CFG) ShowInfo(cmd CmdArgs) (out bool) {
	switch {
	case cmd.V:
		cfg.PrintVersion()
	case cmd.H:
		cfg.PrintHelpInfo()
	default:
		cfg.PrintEnv()
		return false
	}
	return true
}

func (cfg *CFG) PrintEnv() {
	cfg.PrintVersion()
	log.Printf("cmdline: %+v\n", cfg.cmd)
}

func (cfg *CFG) PrintVersion() {
	log.Println(cfg.GetVersion(cfg.EInfo))
}

func (cfg *CFG) PrintHelpInfo() {
	pflag.PrintDefaults()
}
