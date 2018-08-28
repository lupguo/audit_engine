package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mq"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

func main() {
	//mq config
	mqcf := config.CFG.MqConfig

	//conn rabbitmq
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d", mqcf.User, mqcf.Pass, mqcf.Host, mqcf.Port))
	tool.ErrorPanic(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	//channel
	ch, err := conn.Channel()
	tool.ErrorPanic(err, "Failed to open a channel")
	defer ch.Close()

	//soa audit message test
	mq.Publish(*ch, config.AuditQueName["SOA_AUDIT_MSG"], `{"username":"clark","price":100}`)

	log.Println("[*] Send message finish")
}
