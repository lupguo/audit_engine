package task

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

type ConsumeTask struct {
	TkCfg config.CFG
	TkMq  rabbit.MQ
	Que   amqp.Queue
}

//初始化队列任务环境
func (tk *ConsumeTask) Start(qn string, test bool) {
	tool.PrettyPrint(fmt.Sprintf("consume %s task start, do environment init...", qn))

	switch m := config.QueName; qn {
	case m["OBS_PERSON_AUDIT_RESULT"]:
		cfg := tk.TkCfg
		//rabbitmq init conn & channel
		tk.TkMq.Init(cfg.RabbitMq["gb"])
		//create mq
		tk.Que = tk.TkMq.Create(m["SOA_AUDIT_BACK_MSG"])
	}
}

//停止则回收相关信息
func (tk *ConsumeTask) Stop(qn string, test bool) {
	tool.PrettyPrint(fmt.Sprintf("consume %s task stop, do environment clean...", qn))

	switch m := config.QueName; qn {
	case m["OBS_PERSON_AUDIT_RESULT"]:
		tk.TkMq.Close()
	}
}

//针对queue队列名分配任务
func (tk *ConsumeTask) Work(qn string, test bool) (workFn func([]byte)) {
	if test { //test just print message
		return tk.workPrintMessage
	}
	//常规工作队列
	switch m := config.QueName; qn {
	case m["SOA_AUDIT_MSG"]:
		workFn = tk.workAuditMessage
	case m["OBS_RULE_CHANGE_MSG"]:
		workFn = tk.workUpdateRule
	case m["OBS_PERSON_AUDIT_RESULT"]:
		workFn = tk.workUpdateAuditResult
	}

	return workFn
}

//打印 message
func (tk *ConsumeTask) workPrintMessage(msg []byte) {
	log.Println("Working...")
	//tool.PrettyPrint(string(data))
	log.Println("Done")
}

//接收消息审核任务
func (tk *ConsumeTask) workAuditMessage(msg []byte) {
	fmt.Println("Audit Message Task...")
	fmt.Println(msg)
}

//拉取配置任务
func (tk *ConsumeTask) workUpdateRule(msg []byte) {
	fmt.Println("Update Rule Task...")
	fmt.Println(msg)
}

//同步审核结果任务
func (tk *ConsumeTask) workUpdateAuditResult(msg []byte) {
	fmt.Println("Update Rule Result Task...")
	fmt.Println(msg)

	//obs audit result
	var par rabbit.PersonAuditResult
	err := json.Unmarshal(msg, &par)
	//sql update `audit_record` & select * from `audit_record`

	//投递审核结果消息给SOA
	var bk = rabbit.AuditBackMsg{
		SiteCode:    "GB",
		BussUuid:    "13710",
		AuditStatus: 2,
		AuditRemark: "系统审核通过",
		AuditUid:    0,
		AuditUser:   "系统",
		AuditTime:   1535439935871,
	}
	b, err := json.Marshal(bk)
	tool.ErrorLog(err, "publish json marshal fail")

	//msg return
	tk.backMsg(config.QueName["SOA_AUDIT_BACK_MSG"], b)
}

// 响应审核结果消息给到SOA
func (tk *ConsumeTask) backMsg(qn string, msg []byte) {
	//send msg
	tk.TkMq.Publish(tk.Que.Name, msg, 1)
}
