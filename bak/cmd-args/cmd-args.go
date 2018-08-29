package cmd_args

import (
	"github.com/spf13/pflag"
)

type argFlag struct {
	valType   string
	name      string
	shorthand string
	defValue  interface{}
	usage     string
}

var argFlags = []argFlag{
	//help
	{"bool", "version", "v", false, "version info"},
	//test
	{"bool", "test", "t", false, "test model"},
	//mq
	{"string", "publish", nil, "", "publish message from mq"},
	{"string", "consume", nil, "", "consume message from mq"},
	{"string", "queue_name", "q", "", "queue name (soa_audit_msg|soa_audit_back_msg...)"},
	{"string", "data", "D", "", "message data"},
	//config file
	{"string", "config_file", "c", "./config.json", "engine config.json file"},
}

var Args map[string]interface{}

func init() {

	var val, defval interface{}

	for _, v := range argFlags {

		switch v.defValue.(type) {
		case bool:
			defval = false
		case string:
			defval = ""
		}

		switch v.valType {
		case "bool":
			val = pflag.BoolP(v.name, v.shorthand, false, v.usage)
		case "string":
			val = pflag.StringP(v.name, v.shorthand, "", v.usage)
		}
		Args[v.name] = val
	}
}
