package mq

import (
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

//create queue
func Create(ch amqp.Channel, qn string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		qn,
		true,
		false,
		false,
		false,
		nil,
	)
}

//consume work fn bind
func ConsumeBind(ch amqp.Channel, qn string, fn func(data interface{})) {
	//consume
	msgs, err := ch.Consume(
		qn,
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
			fn(d.Body)
			d.Ack(false)
		}
	}()
	log.Printf("[*] Waiting for message. To exit press CTRL+C")
	<-forever
}

//publish message
func Publish(ch amqp.Channel, qn string, data string) {
	//create mq
	q, err := Create(ch, qn)
	tool.ErrorPanic(err, "Failed to declare queue")

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		Body:        []byte(data),
		ContentType: "text/plain",
	})
	tool.ErrorLog(err, "Failed to publish a message")
}
