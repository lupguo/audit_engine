package task

import (
	"encoding/json"
	"fmt"
	"github.com/tkstorm/audit_engine/bucket"
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/rabbit"
	"log"
	"time"
)

//同步审核结果任务
func (tk *ConsumeTask) workUpdateAuditResult(msg []byte) bool {
	log.Println("update audit result task start...")

	db := tk.TkDb.Db

	//obs audit result
	var par rabbit.PersonAuditResult
	err := json.Unmarshal(msg, &par)
	if err != nil {
		log.Println(err, "unmarshal person audit result fail")
		return false
	}

	//sql update `audit_record` & select * from `audit_record`
	sql := "UPDATE audit_message SET audit_status=?, update_time=? WHERE message_id = ?"
	stmt, err := db.Prepare(sql)
	if err != nil {
		log.Println(err, "upd prepare fail")
		return false
	}

	//人工审核状态转换
	adStat := bucket.ObsAudStat[par.Status]
	rst, err := stmt.Exec(adStat, time.Now().Unix(), par.MsgId)
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
	log.Printf("update rows num: %d", rn)

	//send msg to soa
	tk.sendBackMsg(par.MsgId, false)

	log.Println("update audit result task done")
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
	b, err := json.Marshal(bk)
	if err != nil {
		log.Println(err, "marshal result msg data fail")
		return
	}
	log.Printf("audit back msg : %s", b)

	//msg return
	tk.MqGbVh.Publish(config.QueName["SOA_AUDIT_BACK_MSG"], b, 1)
}
