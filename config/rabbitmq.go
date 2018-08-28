package config

type RabbitMqConfig struct {
	Host string
	Port int
	User string
	Pass string
}

//soa 审核消息
type SoaAuditMsg struct {
}

//soa 响应消息
type SoaAuditBackMsg struct {
}

//obs 规则变更消息
type ObsRuleChangeMsg struct {
}

//obs 人工审核结果消息
type ObsPersonAuditResult struct {
}

var AuditQueName = map[string]string{
	"SOA_AUDIT_MSG":           "soa_audit_msg",
	"SOA_AUDIT_BACK_MSG":      "soa_audit_resp_msg",
	"OBS_RULE_CHANGE_MSG":     "obs_rule_change_msg",
	"OBS_PERSON_AUDIT_RESULT": "obs_audit_result_msg",
}
