package bucket

type status map[string]int
type intStatus map[int]int

//10：操作人撤销,20：规则引擎校验，自动通过,21：规则引擎校验，自动拒绝,22：规则全不匹配，自动通过,30：人工审核中,31：人工审核通过,32：人工审核驳回
var MsgStat = status{
	//申请人
	"APPLY_CANCEL": 10,
	//引擎审核
	"EG_AUTO_PASS":    20,
	"EG_AUTO_REJECT":  21,
	"EG_DEFAULT_PASS": 22,
	//obs人工审核
	"OBS_AUDITING":     30,
	"OBS_AUDIT_PASS":   31,
	"OBS_AUDIT_REJECT": 32,
}

//处理状态，1通过，2驳回
var AudStat = intStatus{
	1: MsgStat["OBS_AUDIT_PASS"],
	2: MsgStat["OBS_AUDIT_REJECT"],
}
