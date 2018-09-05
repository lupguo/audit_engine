package task

import (
	"encoding/json"
	"fmt"
	"github.com/tkstorm/audit_engine/rabbit"
	"log"
	"time"
)

//接收消息审核任务
func (tk *ConsumeTask) workAuditMessage(msg []byte) bool {
	log.Println("audit message task start...")

	//审核数据
	var am rabbit.AuditMsg
	err := json.Unmarshal(msg, &am)
	if err != nil {
		log.Println(err, "unmarshal audit message fail")
		return false
	}

	//业务数据
	var bd rabbit.BusinessData
	err = json.Unmarshal([]byte(am.BussData), &bd)
	if err != nil {
		log.Println(err, "unmarshal business data fail")
		return false
	}
	log.Printf("auditMsg: %+v\n", am)
	log.Printf("bussData: %+v\n", bd)

	//hashmap 规则
	hashRuleTypes := tk.GetRuleItems()
	at, ok := hashRuleTypes[am.AuditMark]
	if !ok {
		fmt.Println(am.AuditMark, "hash key not exist")
		return false
	}
	log.Printf("rule type: %+v", at)

	//规则校验(rt)
	matchAction, mrm := RunRuleMatch(&bd, &at)
	log.Println("RunRuleMatch(20：引擎通过,21：引擎拒绝,22：规则全不匹配，自动通过,30：转人工审核)----->", matchAction)

	//自动通过|驳回|转人工审核（写db)
	tk.insertAuditMsg(am, bd, &at, matchAction, mrm)

	log.Println("audit message task done !!")
	return true
}

//审核消息入库
func (tk *ConsumeTask) insertAuditMsg(am rabbit.AuditMsg, bd rabbit.BusinessData, at *AuditType, matchAction int, mrm RuleMatch) {
	db := tk.TkDb.Db

	//检测审核规则是否为空
	if len(at.RuleList) == 0 {
		log.Println("audit rule list is empty")
	}

	//sql
	sql := "INSERT INTO audit_message (" +
		"site_code, rule_id,template_id, audit_sort, audit_mark, audit_name," +
		"business_uuid, business_data, " +
		"create_user,workflow_id,audit_status, module, " +
		"create_time" +
		")" +
		"VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Println(err, "insert into audit_message Prepare fail")
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(
		am.SiteCode,
		mrm.RuleId,
		at.TypeId,
		at.AuditSort,
		at.AuditMark,
		at.TypeTitle,
		am.BussUuid,
		am.BussData,
		am.CreateUser,
		mrm.FlowId,
		matchAction,
		am.Module,
		time.Now().Unix(),
	)
	if err != nil {
		log.Println(err, "insert into audit_message Exec fail")
		return
	}
	lastId, err := result.LastInsertId()
	log.Printf("success insert id: %d", lastId)

	//自动通过或拒绝，发布消息
	if engineReturn := matchAction != AuditStatus[ObsAudit]; engineReturn {
		tk.sendBackMsg(lastId, engineReturn)
	}

}
