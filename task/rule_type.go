package task

import (
	"github.com/tkstorm/audit_engine/config"
	"github.com/tkstorm/audit_engine/mydb"
	"log"
)

//规则哈希表
type AuditTemplate map[string]AuditType

//规则类型
type AuditType struct {
	TypeId    int
	TypeTitle string
	AuditSort int
	AuditMark string
	RuleList  []AuditRule
}

//规则条目
type AuditRule struct {
	RuleId   int
	TypeId   int
	RuleRel  int //1与 2或
	RuleProc int //rule成立后的处理方式，1 系统通过，2 系统驳回，3 转人工审核
	FlowId   int
	Profit   float64
	ItemList []RuleItem
}

type RuleItem struct {
	ItemId      int
	CompareType int
	Field       string
	Operate     string
	Value       string
}

//规则项(compare_type 1:阈值 2:字段）
func GetRuleItems() AuditTemplate {
	log.Println(config.GlobaleCFG)

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
		rows.Scan(&at.TypeId, &at.TypeTitle, &at.AuditSort, &at.AuditMark)
		aTypes = append(aTypes, at)
		//审核ID
		typeIds = append(typeIds, at.TypeId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("aTypes:\n", aTypes)
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
		rows.Scan(&ar.RuleId, &ar.TypeId, &ar.RuleRel, &ar.RuleProc, &ar.FlowId, &ar.Profit)
		aRuls = append(aRuls, ar)
		//规则条目 list
		rids = append(rids, ar.RuleId)

		//分组
		ruleGroups[ar.TypeId] = append(ruleGroups[ar.TypeId], ar)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("aRuls:\n", aRuls)
	//log.Println("ruleGroups:\n", ruleGroups)
	rows.Close()
	stmt.Close()

	//分组+哈希表填充
	hashAuditTemplate := make(AuditTemplate, 10)
	for i, at := range aTypes {
		aTypes[i].RuleList = ruleGroups[at.TypeId]
		hashAuditTemplate[at.AuditMark] = aTypes[i]
	}
	//log.Println("hashAuditTemplate:\n", hashAuditTemplate)

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
		rows.Scan(&k.ItemId, &k.CompareType, &k.Field, &k.Operate, &k.Value)
		items = append(items, k)

		itemGroups[k.ItemId] = append(itemGroups[k.ItemId], k)
	}
	//log.Println("itemGroups:\n", itemGroups)

	//哈希表填充
	for k, t := range hashAuditTemplate {
		for kk, r := range t.RuleList {
			hashAuditTemplate[k].RuleList[kk].ItemList = itemGroups[r.RuleId]
		}
	}

	log.Println("hashAuditTemplate:", hashAuditTemplate)

	return hashAuditTemplate
}