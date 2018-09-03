package task

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/bucket"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/rabbit"
	"github.com/tkstorm/audit_engine/tool"
	"log"
	"time"
)

type ConsumeTask struct {
	TkCfg   config.CFG
	MqGbVh  rabbit.MQ //gb vhost
	MqSoaVh rabbit.MQ //soa vhost
	Que     amqp.Queue
	TkDb    mydb.DbMysql
}

//初始化队列任务环境
func (tk *ConsumeTask) Bootstrap() {
	tool.PrettyPrint("Task Bootstrap...")

	//初始化一个rabbit连接
	cfg := tk.TkCfg
	tk.MqSoaVh.Init(cfg.RabbitMq["soa"])
	tk.MqGbVh.Init(cfg.RabbitMq["gb"])

	//初始化一个db连接
	tk.TkDb.Connect(config.GlobaleCFG.Mysql)
}

//停止则回收相关资源
func (tk *ConsumeTask) Stop() {
	tool.PrettyPrint("Task Clean...")

	tk.MqGbVh.Close()
	tk.TkDb.Close()
}

//基于queue队列名分配工作任务
func (tk *ConsumeTask) GetWork(qn string, test bool) (workFn func([]byte) bool) {
	switch {
	case test:
		workFn = tk.workPrintMessage
	case qn == config.QueName["SOA_AUDIT_MSG"]:
		workFn = tk.workAuditMessage
	case qn == config.QueName["OBS_PERSON_AUDIT_RESULT"]:
		workFn = tk.workUpdateAuditResult
	}
	return workFn
}

//测试仅做打印 message操作
func (tk *ConsumeTask) workPrintMessage(msg []byte) bool {
	log.Println("Working...")
	//tool.PrettyPrint(string(data))
	log.Println("Done")
	return true
}

//接收消息审核任务
func (tk *ConsumeTask) workAuditMessage(msg []byte) bool {
	fmt.Println("Audit Message Task...")
	//获取规则
	var am rabbit.AuditMsg
	err := json.Unmarshal(msg, &am)
	if err != nil {
		tool.ErrorLog(err, "unmarshal audit message fail")
		return false
	}

	var bd rabbit.BusinessData
	err = json.Unmarshal([]byte(am.BussData), &bd)
	if err != nil {
		tool.ErrorLog(err, "unmarshal business data fail")
		return false
	}

	tool.PrettyPrint("AuditMsg:", am)
	tool.PrettyPrint("BussData:", bd)

	//hash map 规则
	hashRuleTypes := GetRuleItems()
	at, ok := hashRuleTypes[am.AuditMark]
	if !ok {
		fmt.Println(am.AuditMark, "hash key not exist")
		return false
	}
	tool.PrettyPrintf("%+v", at)

	//规则校验(rt)
	mr := RunRuleMatch(&bd, &at)
	fmt.Println("RunRuleMatch----->:", mr)

	//自动通过|驳回|转人工审核（写db)
	tk.insertAuditMsg(am, bd, &at, mr)

	tool.PrettyPrint("Audit Message Task Done !!")
	return true
}

//审核消息入库
func (tk *ConsumeTask) insertAuditMsg(am rabbit.AuditMsg, bd rabbit.BusinessData, at *AuditType, mr int) {

	db := tk.TkDb.Db

	//sql
	sql := "INSERT INTO audit_message (" +
		"site_code,template_id, audit_sort, audit_mark, audit_name," +
		"business_uuid, business_data, " +
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

	//检测审核规则是否为空
	flowId := 0
	if len(at.RuleList) > 0 {
		tool.ErrorLogP("Audit rule list is empty")
		flowId = at.RuleList[0].FlowId
	}

	result, err := stmt.Exec(
		am.SiteCode,
		at.TypeId,
		at.AuditSort,
		at.AuditMark,
		at.TypeTitle,
		am.BussUuid,
		am.BussData,
		am.CreateUser,
		flowId,
		mr,
		time.Now().Unix(),
	)
	if err != nil {
		tool.ErrorLog(err, "insert into audit_message Exec fail")
		return
	}
	lastId, err := result.LastInsertId()
	tool.PrettyPrintf("Success insert id: %d", lastId)

	//自动通过或拒绝，发布消息
	if engineReturn := mr != AuditStatus[ObsAudit]; engineReturn {
		tk.sendBackMsg(lastId, engineReturn)
	}

}

//同步审核结果任务
func (tk *ConsumeTask) workUpdateAuditResult(msg []byte) bool {
	tool.PrettyPrint("Update Rule Result Task...")

	db := tk.TkDb.Db

	//obs audit result
	var par rabbit.PersonAuditResult
	err := json.Unmarshal(msg, &par)
	if err != nil {
		tool.ErrorLog(err, "unmarshal person audit result fail")
		return false
	}

	//sql update `audit_record` & select * from `audit_record`
	sql := "UPDATE audit_message SET audit_status=? WHERE message_id = ?"
	stmt, err := db.Prepare(sql)
	if err != nil {
		tool.ErrorLog(err, "upd prepare fail")
		return false
	}

	//人工审核状态转换
	adStat := bucket.ObsAudStat[par.Status]
	rst, err := stmt.Exec(adStat, par.MsgId)
	if err != nil {
		tool.ErrorLog(err, "upd audit status exec fail(stmt.exec)")
		return false
	}
	stmt.Close()

	rn, err := rst.RowsAffected()
	if err != nil || rn == 0 {
		tool.ErrorLog(err, "upd audit status fail(update none row)")
		return false
	}
	tool.PrettyPrintf("Success update rows: %d", rn)

	//send msg to soa
	tk.sendBackMsg(par.MsgId, false)

	tool.PrettyPrint("Done")
	return true
}

func (tk *ConsumeTask) sendBackMsg(msgId int64, engineReturn bool) {
	db := tk.TkDb.Db

	//select info & return to soa_back_msg
	sql := ""
	if engineReturn {
		sql = fmt.Sprintf(`
SELECT
  m.site_code,
  m.business_uuid,
  m.audit_mark,
  m.audit_status,
	"%s" as audit_explain,
	%d as user_id,
	%d as create_time
FROM audit_message as m
WHERE m.message_id = ? ORDER BY message_id desc LIMIT 1;`,
			"系统自动审核",
			0,
			time.Now().Unix(),
		)
	} else {
		sql = `
SELECT
  m.site_code,
  m.business_uuid,
  m.audit_mark,
  m.audit_status,
  mr.audit_explain,
  mr.user_id,
  mr.create_time
FROM audit_record  AS mr LEFT JOIN audit_message as m USING(message_id)
WHERE m.message_id = ? ORDER BY message_id desc LIMIT 1;`
	}

	rows := db.QueryRow(sql, msgId)
	fmt.Println(sql, msgId)

	var bk rabbit.AuditBackMsg
	err := rows.Scan(
		&bk.SiteCode,
		&bk.BussUuid,
		&bk.AuditMark,
		&bk.AuditStatus,
		&bk.AuditRemark,
		&bk.AuditUid,
		&bk.AuditTime,
	)
	if err != nil {
		tool.ErrorLog(err, "rows scan fail")
		return
	}
	bk.AuditStatus = bucket.SoaAudStat[bk.AuditStatus]
	tool.PrettyPrintf("audit back msg : %+v", bk)
	b, err := json.Marshal(bk)
	if err != nil {
		tool.ErrorLog(err, "marshal result msg data fail")
		return
	}

	//msg return
	tk.MqGbVh.Publish(config.QueName["SOA_AUDIT_BACK_MSG"], b, 1)
}
