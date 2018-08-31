package rabbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

type Config struct {
	Host  string
	Port  int
	User  string
	Pass  string
	Vhost string
}

type MQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (mq *MQ) Close() {
	mq.conn.Close()
	mq.ch.Close()
}

//队列初始化
func (mq *MQ) Init(mqcf Config) {
	var err error
	//conn
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", mqcf.User, mqcf.Pass, mqcf.Host, mqcf.Port, mqcf.Vhost)
	tool.PrettyPrint(url)
	mq.conn, err = amqp.Dial(url)
	tool.ErrorPanic(err, "Failed to connect to RabbitMQ")

	//channel
	mq.ch, err = mq.conn.Channel()
	tool.ErrorPanic(err, "Failed to open a channel")
}

//队列创建
func (mq *MQ) Create(qn string) amqp.Queue {
	durable, autoDelete := true, false
	if qn == "" {
		durable = false
		autoDelete = true
	}
	q, err := mq.ch.QueueDeclare(
		qn,
		durable,
		autoDelete,
		false,
		false,
		nil,
	)
	tool.ErrorPanic(err, "Failed to declare queue")
	return q
}

//队列消费程序绑定
func (mq *MQ) ConsumeBind(qn string, fn func([]byte), noAck bool, istest bool) {
	//consume resister
	msgs, err := mq.ch.Consume(
		qn,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	tool.ErrorLog(err, "Failed to register a consumer")

	//consume work
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			tool.PrettyPrintf("Received a message: %s", d.Body)
			fn(d.Body)
			if !noAck {
				d.Ack(false)
			}
		}
	}()
	log.Printf("[*] Waiting for message. To exit press CTRL+C")
	<-forever
}

//队列发布消息
func (mq *MQ) Publish(qn string, data []byte, n int) {
	//publish msg
	msg := amqp.Publishing{
		Body:        data,
		ContentType: "text/plain",
	}

	for i := 0; i < n; i++ {
		err := mq.ch.Publish("", qn, false, false, msg)
		tool.ErrorLog(err, "Failed to publish a message")
		tool.PrettyPrint("Send message finish")
	}
}
