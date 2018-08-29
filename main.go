package main

import (
	"encoding/json"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

var (
	cmd  config.CmdArgs
	cfg  config.CFG
	mq   rabbit.MQ
	rbmq rabbit.Config
)

func main() {
	//cmd args parse from cmdline
	cmd.Parse()

	//config init from cmdline args
	cfg.Init(cmd)

	//show message
	switch {
	case cmd.V:
		cfg.PrintVersion()
		return
	case cmd.H:
		cfg.PrintHelpInfo()
		return
	default:
		cfg.PrintEnv()
	}

	//rabbitmq init conn & channel
	mq.Init(cfg.RabbitMq)
	defer mq.Close()

	//create mq
	q := mq.Create(cmd.QName)
	tool.PrettyPrint("deal with queue:", q.Name)

	//publish or consume
	if cmd.Pub {
		tool.PrettyPrint("message publish...")
		mq.Publish(q.Name, msgData(q.Name), cmd.RepNumber)
	} else {
		tool.PrettyPrint("message consume...")
		mq.ConsumeBind(q.Name, func() func([]byte) bool {
			if cmd.T {
				return printMessage
			} else {
				return doAuditWork
			}
		}())
	}
}

//print message
func printMessage(data []byte) bool {
	log.Println("Working...")
	//tool.PrettyPrint(string(data))
	log.Println("Done")
	return true
}

//消息解码，引擎分析，DB入库
func doAuditWork(data []byte) bool {
	log.Println("Begin audit...")

	//消费auditMsg
	var m rabbit.AuditMsg
	err := json.Unmarshal(data, &m)
	tool.ErrorPanic(err, "Json unmarshal failed")

	tool.PrettyPrint(m)
	log.Println("Done")

	return true
}

// select queue and prepare message data
func msgData(qn string) []byte {
	if cmd.MsgData != "" {
		return []byte(cmd.MsgData)
	}

	switch qn {
	case config.QueName["SOA_AUDIT_BACK_MSG"]:
		return getAuditBackMsg()
	case config.QueName["SOA_AUDIT_MSG"]:
		return getAuditMsg()
	case config.QueName["OBS_RULE_CHANGE_MSG"]:
		return []byte(`{"action":"upd|del|add","templete_id":1}`)
	case config.QueName["OBS_PERSON_AUDIT_RESULT"]:
		return []byte(`{"message_id":"1","status":"2"}`)
	default:
		return []byte("Heyman, Cool")
	}
}

//审核响应消息
func getAuditBackMsg() []byte {
	var msg = rabbit.AuditBackMsg{
		SiteCode:    "GB",
		BussUuid:    "13710",
		AuditStatus: 2,
		AuditRemark: "系统审核通过",
		AuditUid:    0,
		AuditUser:   "系统",
		AuditTime:   1535439935871,
	}

	b, err := json.Marshal(msg)
	tool.ErrorLog(err, "publish json marshal fail")
	return b
}

//审核消息
func getAuditMsg() []byte {
	return []byte(`{"auditMark":"goods-price-check","businessData":"{\"calculatePrice\":52.02,\"catId\":11286,\"changeType\":1,\"chargePrice\":59.00000,\"goodSn\":\"YL4225902\",\"pipelineCode\":\"GB\",\"rate\":3.68,\"sysLabelId\":-1,\"virWhCode\":\"1433363\"}","businessUuid":"13710","createTime":1535427621607,"createUser":"huang","createUserId":0,"module":"goods","siteCode":"GB"}`)
}
