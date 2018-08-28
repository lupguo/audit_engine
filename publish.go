package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

func main() {
	//mq
	mqcf := config.Engine.MqConfig

	//conn rabbitmq
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d", mqcf.User, mqcf.Pass, mqcf.Host, mqcf.Port))
	tool.ErrorPanic(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	//channel
	ch, err := conn.Channel()
	tool.ErrorPanic(err, "Failed to open a channel")
	defer ch.Close()

	//create mq
	qn := config.AuditQueName["SOA_AUDIT_MSG"]
	q, err := createMq(*ch, qn)
	tool.ErrorPanic(err, "Failed to declare queue")

	//publish msg
	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		Body:        []byte(`{"username":"clark","price":100}`),
		ContentType: "text/plain",
	})
	tool.ErrorLog(err, "Failed to publish a message")

	log.Println("Send message finish")
}

//创建mq队列
func createMq(ch amqp.Channel, qn string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		qn,
		true,
		false,
		false,
		false,
		nil,
	)
}
