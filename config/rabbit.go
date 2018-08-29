package config

import "github.com/tkstorm/audit_engine/rabbit"

var QueName = rabbit.QueueName{
	"SOA_AUDIT_MSG":           "soa_audit_msg",
	"SOA_AUDIT_BACK_MSG":      "soa_audit_resp_msg",
	"OBS_RULE_CHANGE_MSG":     "obs_rule_change_msg",
	"OBS_PERSON_AUDIT_RESULT": "obs_audit_result_msg",
}
