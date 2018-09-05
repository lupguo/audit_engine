package task

import (
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/rabbit"
	"log"
)

type ConsumeTask struct {
	TkCfg   config.CFG
	MqGbVh  rabbit.MQ //gb vhost
	MqSoaVh rabbit.MQ //soa vhost
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
	tk.TkDb.Connect(cfg.Mysql)
}

//停止则回收相关资源
func (tk *ConsumeTask) Stop() {
	log.Println("task clean...")

	tk.MqSoaVh.Close()
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
