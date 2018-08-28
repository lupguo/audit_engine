package test

import (
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mq"
)

//soa审核消息
func PostAuditMsg(ch amqp.Channel) {
	//publish msg
	mq.Publish(ch, config.AuditQueName["SOA_AUDIT_MSG"], `{"username":"clark","price":100}`)
}
