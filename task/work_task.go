package task

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/tkstorm/audit_engine/bucket"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/rabbit"
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
	log.Println("task bootstrap...")

	//初始化一个rabbit连接
	cfg := tk.TkCfg
	tk.MqSoaVh.Init(cfg.RabbitMq["soa"])
	tk.MqGbVh.Init(cfg.RabbitMq["gb"])

	//初始化一个db连接
	tk.TkDb.Connect(config.GlobaleCFG.Mysql)
}

//停止则回收相关资源
func (tk *ConsumeTask) Stop() {
	log.Println("task clean...")

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
	log.Println("working...")
	//log.Println(string(data))
	log.Println("done")
	return true
}

//接收消息审核任务
func (tk *ConsumeTask) workAuditMessage(msg []byte) bool {
	fmt.Println("audit message task...")
	//获取规则
	var am rabbit.AuditMsg
	err := json.Unmarshal(msg, &am)
	if err != nil {
		log.Println(err, "unmarshal audit message fail")
		return false
	}

	var bd rabbit.BusinessData
	err = json.Unmarshal([]byte(am.BussData), &bd)
	if err != nil {
		log.Println(err, "unmarshal business data fail")
		return false
	}

	log.Println("auditMsg:", am)
	log.Println("bussData:", bd)

	//hash map 规则
	hashRuleTypes := GetRuleItems()
	at, ok := hashRuleTypes[am.AuditMark]
	if !ok {
		fmt.Println(am.AuditMark, "hash key not exist")
		return false
	}
	log.Printf("%+v", at)

	//规则校验(rt)
	matchAction, mrm := RunRuleMatch(&bd, &at)
	log.Println("RunRuleMatch----->", matchAction)

	//自动通过|驳回|转人工审核（写db)
	tk.insertAuditMsg(am, bd, &at, matchAction, mrm)

	log.Println("Audit Message Task Done !!")
	return true
}

//审核消息入库
func (tk *ConsumeTask) insertAuditMsg(am rabbit.AuditMsg, bd rabbit.BusinessData, at *AuditType, matchAction int, mrm RuleMatch) {
	db := tk.TkDb.Db

	//检测审核规则是否为空
	if len(at.RuleList) == 0 {
		log.Println("Audit rule list is empty")
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
	log.Printf("Success insert id: %d", lastId)

	//自动通过或拒绝，发布消息
	if engineReturn := matchAction != AuditStatus[ObsAudit]; engineReturn {
		tk.sendBackMsg(lastId, engineReturn)
	}

}

//同步审核结果任务
func (tk *ConsumeTask) workUpdateAuditResult(msg []byte) bool {
	log.Println("Update Rule Result Task...")

	db := tk.TkDb.Db

	//obs audit result
	var par rabbit.PersonAuditResult
	err := json.Unmarshal(msg, &par)
	if err != nil {
		log.Println(err, "unmarshal person audit result fail")
		return false
	}

	//sql update `audit_record` & select * from `audit_record`
	sql := "UPDATE audit_message SET audit_status=? WHERE message_id = ?"
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Println(err, "upd prepare fail")
		return false
	}

	//人工审核状态转换
	adStat := bucket.ObsAudStat[par.Status]
	rst, err := stmt.Exec(adStat, par.MsgId)
	if err != nil {
		log.Println(err, "upd audit status exec fail(stmt.exec)")
		return false
	}
	stmt.Close()

	rn, err := rst.RowsAffected()
	if err != nil || rn == 0 {
		log.Println(err, "upd audit status fail(update none row)")
		return false
	}
	log.Printf("Success update rows: %d", rn)

	//send msg to soa
	tk.sendBackMsg(par.MsgId, false)

	log.Println("Done")
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
	"%s" as username,
	%d as create_time
FROM audit_message as m
WHERE m.message_id = ? ORDER BY message_id desc LIMIT 1;`,
			"系统自动审核",
			0,
			"系统",
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
  mr.username,
  mr.create_time
FROM audit_record  AS mr LEFT JOIN audit_message as m USING(message_id)
WHERE m.message_id = ? ORDER BY message_id desc LIMIT 1;`
	}

	rows := db.QueryRow(sql, msgId)
	//fmt.Println(sql, msgId)

	var bk rabbit.AuditBackMsg
	err := rows.Scan(
		&bk.SiteCode,
		&bk.BussUuid,
		&bk.AuditMark,
		&bk.AuditStatus,
		&bk.AuditRemark,
		&bk.AuditUid,
		&bk.AuditUser,
		&bk.AuditTime,
	)
	if err != nil {
		log.Println(err, "rows scan fail")
		return
	}
	bk.AuditStatus = bucket.SoaAudStat[bk.AuditStatus]
	log.Printf("audit back msg : %+v", bk)
	b, err := json.Marshal(bk)
	if err != nil {
		log.Println(err, "marshal result msg data fail")
		return
	}

	//msg return
	tk.MqGbVh.Publish(config.QueName["SOA_AUDIT_BACK_MSG"], b, 1)
}
