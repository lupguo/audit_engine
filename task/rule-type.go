package task

import (
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"github.com/tkstorm/audit_engine/tool"
	"log"
)

//规则哈希表
type AuditTemplate map[string]AuditType

//规则类型
type AuditType struct {
	typeId    int
	typeTitle string
	auditSort int
	auditMark string
	//typeDesc  string
	ruleList []AuditRule
}

//规则条目
type AuditRule struct {
	ruleId    int
	typeId    int
	ruleRel   int //1与 2或
	ruleProc  int //rule成立后的处理方式，1 系统通过，2 系统驳回，3 转人工审核
	flowId    int
	profit    float64
	ruleItems []RuleItem
}

type RuleItem struct {
	itemId      int
	compareType int
	field       string
	operate     string
	value       string
}

//规则项(compare_type 1:阈值 2:字段）
func GetRuleItems() AuditTemplate {
	tool.PrettyPrint(config.GlobaleCFG)

	var dbMysql mydb.DbMysql
	dbcf := config.GlobaleCFG.Mysql
	dbMysql.Connect(dbcf)
	defer dbMysql.Close()

	//---------审核类型
	sql := `select id, title, sort,audit_mark from audit_template;`
	rows, err := dbMysql.Db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}

	var aTypes []AuditType
	var typeIds []interface{}
	for rows.Next() {
		var at AuditType
		rows.Scan(&at.typeId, &at.typeTitle, &at.auditSort, &at.auditMark)
		aTypes = append(aTypes, at)
		//审核ID
		typeIds = append(typeIds, at.typeId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//tool.PrettyPrint("aTypes:\n", aTypes)
	rows.Close()

	//----------规则条目
	stmt, err := dbMysql.Db.Prepare("select id ,template_id, items_relation, process_type, workflow_id, base_profit_margin " +
		"from audit_rule WHERE template_id IN (" + mydb.Concat(typeIds) + ") " +
		"ORDER BY sort ASC;")
	if err != nil {
		log.Fatal(err)
	}
	rows, err = stmt.Query(typeIds...)
	if err != nil {
		log.Fatal(err)
	}

	var aRuls []AuditRule
	var rids []interface{}
	var ar AuditRule
	ruleGroups := make(map[int][]AuditRule, len(aTypes))
	for rows.Next() {
		rows.Scan(&ar.ruleId, &ar.typeId, &ar.ruleRel, &ar.ruleProc, &ar.flowId, &ar.profit)
		aRuls = append(aRuls, ar)
		//规则条目 list
		rids = append(rids, ar.ruleId)

		//分组
		ruleGroups[ar.typeId] = append(ruleGroups[ar.typeId], ar)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//tool.PrettyPrint("aRuls:\n", aRuls)
	//tool.PrettyPrint("ruleGroups:\n", ruleGroups)
	rows.Close()
	stmt.Close()

	//分组+哈希表填充
	hashAuditTemplate := make(AuditTemplate, 10)
	for i, at := range aTypes {
		aTypes[i].ruleList = ruleGroups[at.typeId]
		hashAuditTemplate[at.auditMark] = aTypes[i]
	}
	//tool.PrettyPrint("hashAuditTemplate:\n", hashAuditTemplate)

	//--------比较项
	sql = "select rule_id,compare_type,field,operation,value from audit_rule_item WHERE rule_id IN (" + mydb.Concat(rids) + ")"
	stmt, err = dbMysql.Db.Prepare(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err = stmt.Query(rids...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var items []RuleItem
	itemGroups := make(map[int][]RuleItem, len(aRuls))
	for rows.Next() {
		var k RuleItem
		rows.Scan(&k.itemId, &k.compareType, &k.field, &k.operate, &k.value)
		items = append(items, k)

		itemGroups[k.itemId] = append(itemGroups[k.itemId], k)
	}
	//tool.PrettyPrint("itemGroups:\n", itemGroups)

	//哈希表填充
	for k, t := range hashAuditTemplate {
		for kk, r := range t.ruleList {
			hashAuditTemplate[k].ruleList[kk].ruleItems = itemGroups[r.ruleId]
		}
	}

	tool.PrettyPrint("hashAuditTemplate:\n", hashAuditTemplate)

	return hashAuditTemplate
}
