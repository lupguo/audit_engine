package config

import (
	"github.com/spf13/pflag"
	"strings"
)

type CmdArgs struct {
	//engine
	V bool
	H bool
	//test
	T     bool
	NoAck bool
	//config file
	Cfg string
	//mq
	QName     string
	Pub       bool
	Cus       bool
	MsgData   string
	RepNumber int
}

func (cmd *CmdArgs) Parse() {
	//engine
	pflag.BoolVarP(&cmd.V, "version", "v", false, "version info")
	pflag.BoolVarP(&cmd.H, "help", "h", false, "show this message")
	//test
	pflag.BoolVarP(&cmd.T, "test", "t", false, `test model for msg publish or consume`)
	pflag.BoolVar(&cmd.NoAck, "no_ack", false, `do not return ack the message to the Broker`)
	//get config file
	pflag.StringVarP(&cmd.Cfg, "config_file", "c", "./config.json", "config file")
	//mq
	pflag.BoolVar(&cmd.Pub, "publish", false, "run publish message")
	pflag.BoolVar(&cmd.Cus, "consume", false, "run consume message")
	pflag.StringVarP(&cmd.QName, "queue", "q", "", `message queue name(`+availQueue()+`)`)
	pflag.StringVarP(&cmd.MsgData, "msg_data", "D", "", `message data used for publish`)
	pflag.IntVarP(&cmd.RepNumber, "rep_number", "n", 1, `message data send rep_number time`)
	pflag.Parse()
}

func availQueue() string {
	var q []string
	for _, v := range QueName {
		q = append(q, v)
	}
	return strings.Join(q, "|")
}
