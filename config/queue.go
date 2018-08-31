package config

import "github.com/tkstorm/audit_engine/rabbit"

var QueName = rabbit.QueueName{
	"SOA_AUDIT_BACK_MSG":      "auditResult_SOA_GOODS",
	"SOA_AUDIT_MSG":           "auditMessage_GB",
	"OBS_RULE_CHANGE_MSG":     "obsRuleChange_GB",
	"OBS_PERSON_AUDIT_RESULT": "obsAuditResult_GB",
}
