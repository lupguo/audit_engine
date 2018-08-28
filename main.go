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

	//create mq
	q, err := mq.Create(*ch, config.AuditQueName["SOA_AUDIT_MSG"])
	tool.ErrorPanic(err, "Failed to declare queue")

	//consume
	mq.ConsumeBind(*ch, q.Name, doWork)
}

//worker
func doWork(data interface{}) {
	log.Println("begin audit...")
	log.Println("Done")
}
