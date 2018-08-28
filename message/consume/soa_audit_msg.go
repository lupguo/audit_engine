package consume

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

//接收soa请求消息 receiveMQ
func ReceiveSoaAuditData(mqcf config.RabbitMqConfig) {
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
	q, err := ch.QueueDeclare(
		qn,
		true,
		false,
		false,
		false,
		nil,
	)
	tool.ErrorPanic(err, "Failed to declare queue")

	//consume
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	forever := make(chan bool)
	go func() {
		tool.ErrorLog(err, "Failed to register a consumer")
		for d := range msgs {
			log.Printf("Received a message: %s\n", d.Body)
			log.Println("Done")
			doWork(d.Body)
			d.Ack(false)
		}
	}()
	log.Printf(" [*] Waiting for message. To exit press CTRL+C")
	<-forever
}

//消费规则校验
func doWork(data interface{}) {
	fmt.Printf("%s", data)
}
