package task

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
	"log"
	"time"
)

type ConsumeTask struct {
	TkCfg config.CFG
	TkMq  rabbit.MQ
	Que   amqp.Queue
	TkDb  mydb.DbMysql
}

//初始化队列任务环境
func (tk *ConsumeTask) Start(qn string, test bool) {
	tool.PrettyPrint(fmt.Sprintf("consume %s task start, do environment init...", qn))

	switch m := config.QueName; qn {
	case m["OBS_RULE_CHANGE_MSG"]:
	case m["SOA_AUDIT_MSG"]:
		//初始化一个db连接
		tk.TkDb.Connect(config.GlobaleCFG.Mysql)
	case m["OBS_PERSON_AUDIT_RESULT"]:
		//初始化一个rabbit连接
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

	//获取规则
	var s rabbit.AuditMsg
	err := json.Unmarshal(msg, &s)
	if err != nil {
		tool.ErrorLog(err, "unmarshal audit message fail")
		return
	}

	var t rabbit.BusinessData
	err = json.Unmarshal([]byte(s.BussData), &t)
	if err != nil {
		tool.ErrorLog(err, "unmarshal business data fail")
		return
	}

	tool.PrettyPrint("AuditMsg:", s)
	tool.PrettyPrint("BussData:", t)

	//hash map 规则
	hashRuleTypes := GetRuleItems()
	at, ok := hashRuleTypes[s.AuditMark]
	if !ok {
		fmt.Println(s.AuditMark, "hash key not exist")
		return
	}
	tool.PrettyPrintf("%+v", at)

	//规则校验(rt)

	//自动通过|驳回|转人工审核（写db)
	tk.insertAuditMsg(s, t, at)

	log.Println("Done")
}

func (tk *ConsumeTask) insertAuditMsg(ad rabbit.AuditMsg, bd rabbit.BusinessData, at AuditType) {
	db := tk.TkDb.Db

	//sql
	sql := "INSERT INTO audit_message (" +
		"site_code,template_id, audit_mark, audit_name," +
		"audit_desc,business_uuid, business_data, " +
		"create_user,workflow_id,audit_status, " +
		"create_time" +
		")" +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		tool.ErrorLog(err, "insert into audit_message Prepare fail")
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		ad.SiteCode,
		at.typeId,
		at.auditMark,
		at.typeTitle,
		at.typeDesc,
		ad.BussUuid,
		ad.BussData,
		ad.CreateUser,
		at.ruleList[0].flowId,
		"30",
		time.Now().Unix(),
	)
	if err != nil {
		tool.ErrorLog(err, "insert into audit_message Exec fail")
		return
	}
	lastId, err := result.LastInsertId()
	tool.PrettyPrintf("Success insert id: %d", lastId)
}

//拉取配置任务
func (tk *ConsumeTask) workUpdateRule(msg []byte) {
	tool.PrettyPrint("Update Rule Task...")
	GetRuleItems()
	tool.PrettyPrint("Done")
}

//同步审核结果任务
func (tk *ConsumeTask) workUpdateAuditResult(msg []byte) {
	tool.PrettyPrint("Update Rule Result Task...")

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

	tool.PrettyPrint("Done")
}

// 响应审核结果消息给到SOA
func (tk *ConsumeTask) backMsg(qn string, msg []byte) {
	//send msg
	tk.TkMq.Publish(tk.Que.Name, msg, 1)
}
