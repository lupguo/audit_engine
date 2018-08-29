package consume

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
)

var mqcf = config.CFG.MqConfig

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
	q, err := rabbit.Create(*ch, config.AuditQueName["SOA_AUDIT_MSG"])
	tool.ErrorPanic(err, "Failed to declare queue")

	//consume
	rabbit.ConsumeBind(*ch, q.Name, doWork)
}

//worker
func doWork(data interface{}) {
	fmt.Printf("%s", data)
}
