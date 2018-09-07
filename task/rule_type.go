package task

import (
	"github.com/tkstorm/audit_engine/mydb"
	"log"
)

//规则哈希表
type AuditTypeList map[string]AuditType

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
	RuleId      int
	CompareType int
	Field       string
	Operate     string
	Value       string
}

var z AuditTypeList

//规则项(compare_type 1:阈值 2:字段）
func (tk *ConsumeTask) GetRuleItems() AuditTypeList {

	log.Print("sql get rule items!!")

	db := tk.TkDb.Db

	//---------审核类型
	sql := `select id, title, sort,audit_mark from audit_template;`
	rows, err := db.Query(sql)
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
	stmt, err := db.Prepare("select id ,template_id, items_relation, process_type, workflow_id, base_profit_margin " +
		"from audit_rule WHERE template_id IN (" + mydb.Concat(typeIds) + ") " +
		"ORDER BY sort ASC;")
	if err != nil {
		log.Fatal(err)
	}
	rows, err = stmt.Query(typeIds...)
	if err != nil {
		log.Fatal(err)
	}

	var aRules []AuditRule
	var rids []interface{}
	var ar AuditRule
	ruleGroups := make(map[int][]AuditRule, len(aTypes))
	for rows.Next() {
		rows.Scan(&ar.RuleId, &ar.TypeId, &ar.RuleRel, &ar.RuleProc, &ar.FlowId, &ar.Profit)
		aRules = append(aRules, ar)
		//规则条目 list
		rids = append(rids, ar.RuleId)

		//分组
		ruleGroups[ar.TypeId] = append(ruleGroups[ar.TypeId], ar)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("aRules:\n", aRules)
	//log.Println("ruleGroups:\n", ruleGroups)
	rows.Close()
	stmt.Close()

	//分组+哈希表填充
	hashAuditTypeList := make(AuditTypeList, 10)
	for i, at := range aTypes {
		aTypes[i].RuleList = ruleGroups[at.TypeId]
		hashAuditTypeList[at.AuditMark] = aTypes[i]
	}
	//log.Println("hashAuditTypeList:\n", hashAuditTypeList)

	//--------比较项
	sql = "select id, rule_id,compare_type,field,operation,value from audit_rule_item WHERE rule_id IN (" + mydb.Concat(rids) + ")"
	stmt, err = db.Prepare(sql)
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
	itemGroups := make(map[int][]RuleItem, len(aRules))
	for rows.Next() {
		var k RuleItem
		rows.Scan(&k.ItemId, &k.RuleId, &k.CompareType, &k.Field, &k.Operate, &k.Value)
		items = append(items, k)
		itemGroups[k.RuleId] = append(itemGroups[k.RuleId], k)
	}
	//log.Println("itemGroups:\n", itemGroups)

	//哈希表填充
	for k, t := range hashAuditTypeList {
		for kk, r := range t.RuleList {
			hashAuditTypeList[k].RuleList[kk].ItemList = itemGroups[r.RuleId]
		}
	}

	log.Printf("hashAuditTypeList: %+v\n", hashAuditTypeList)

	return hashAuditTypeList
}
