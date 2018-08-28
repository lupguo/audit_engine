package main

import (
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/message/consume"
)

func main() {

	egcf := config.Engine

	//start consume soa
	consume.ReceiveSoaAuditData(egcf.MqConfig)

	//start soa consume engine
}
