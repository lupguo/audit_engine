package tool

import (
	"encoding/json"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/rabbit"
)

// select queue and prepare message data
func SimulateData(qn string, cmd *config.CmdArgs) []byte {
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
		return []byte(`{"message_id":1,"status": 2}`)
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
	FatalLog(err, "publish json marshal fail")
	return b
}

//审核消息
func getAuditMsg() []byte {
	return []byte(`{"auditMark":"goods-price-check","bussData":"{\"calculatePrice\":52.02,\"catId\":11286,\"changeType\":1,\"chargePrice\":59.00000,\"goodSn\":\"YL4225902\",\"pipelineCode\":\"GB\",\"rate\":3.68,\"sysLabelId\":-1,\"virWhCode\":\"1433363\"}","bussUuid":"13710","createTime":1535427621607,"createUser":"huang","createUserId":0,"module":"goods","siteCode":"GB"}`)
}
