package account

import (
	"bytes"
	"fmt"
	"strings"
	"syscall"
	"text/template"
	"time"
)

const (
	QueryWildcard = "%"
)

type Type int

const (
	All Type = iota
	Asset
	Bank
	Cash
	Credit
	Expense
	Income
	Liability
)

var typeStrings []string = []string{
	"ALL",
	"ASSET",
	"BANK",
	"CASH",
	"CREDIT",
	"EXPENSE",
	"INCOME",
	"LIABILITY",
}

// Reference: https://lists.gnucash.org/pipermail/gnucash-user/2014-December/057344.html
const (
	query = `WITH RECURSIVE tree (
                     guid,
                     parent_guid,
                     name,
                     name_tree,
                     name_tabs,
                     account_type,
                     depth
                 ) AS (
                     SELECT guid,
                            parent_guid,
                            name,
                            '' || name AS name_tree,
                            '' AS name_tabs,
                            account_type,
                            0 AS depth
                     FROM accounts
                     WHERE parent_guid IS NULL
                     AND NAME <> 'Template Root'
                     UNION ALL
                     SELECT a.guid,
                            a.parent_guid,
                            a.name,
                            tree.name_tree || ':' || a.name AS name_tree,
                            substr('.................................',1,depth*2) || a.name AS name_tabs,
                            a.account_type,
                            depth + 1 AS depth
                     FROM tree
                     JOIN accounts a
                     ON tree.guid = a.parent_guid
                 )
                 SELECT '{{.Type}}' AS account_type,
                     datetime({{.Time1}}, 'unixepoch', 'localtime') AS start_date,
                     datetime({{.Time2}}, 'unixepoch', 'localtime') AS end_date,
                     tr.depth,
                     tr.name AS depth_name,
                     tr.name_tree AS full_name,
                     COALESCE(({{.SubTxnCount}}), 0) AS transaction_count,
                     COALESCE(({{.SubDebits}}), 0) AS debits,
                     COALESCE(({{.SubCredits}}), 0) AS credits,
                     COALESCE(({{.SubBalance}}), 0) AS balance
                 FROM tree tr
                 WHERE tr.depth = {{.Depth}}
                 AND tr.account_type IN ({{.TypeValues}})
                 AND tr.name_tree LIKE '{{.Name}}'
                 AND balance <> 0
                 ORDER BY tr.name_tree, balance DESC;`

	txn = `SELECT {{.Statement}} AS value
               FROM tree tr1
               LEFT JOIN splits s
               ON tr1.guid = s.account_guid
               {{.Condition}}
               LEFT JOIN transactions tx
               ON s.tx_guid = tx.guid
               WHERE tr1.depth >= tr.depth
               AND tr1.name_tree LIKE tr.name_tree || '%'
               AND strftime('%s', tx.post_date) BETWEEN "{{.Time1}}" AND "{{.Time2}}"`
)

type Row struct {
	Type      string  `csv:"account_type"      db:"account_type"`
	StartDate string  `csv:"start_date"        db:"start_date"`
	EndDate   string  `csv:"end_date"          db:"end_date"`
	Depth     int     `csv:"depth"             db:"depth"`
	DepthName string  `csv:"depth_name"        db:"depth_name"`
	FullName  string  `csv:"full_name"         db:"full_name"`
	TxnCount  int     `csv:"transaction_count" db:"transaction_count"`
	Debits    float64 `csv:"debits"            db:"debits"`
	Credits   float64 `csv:"credits"           db:"credits"`
	Balance   float64 `csv:"balance"           db:"balance"`
}

type txnData struct {
	Statement string
	Condition string
	Time1     int64
	Time2     int64
}

type qryData struct {
	Depth       int
	Name        string
	SubBalance  string
	SubCredits  string
	SubDebits   string
	SubTxnCount string
	Type        string
	TypeValues  string
	Time1       int64
	Time2       int64
}

func newTxnValue(statement, condition string, time1, time2 time.Time) string {
	qry := ""
	if t, err := template.New("txn").Parse(txn); err == nil {
		var buf bytes.Buffer
		input := txnData{
			Statement: statement,
			Condition: condition,
			Time1:     time1.Unix(),
			Time2:     time2.Unix(),
		}
		if err = t.Execute(&buf, input); err == nil {
			qry = buf.String()
		}
	}
	return qry
}

func NewQuery(typ Type, name string, depth int, time1, time2 time.Time) string {
	qry := ""
	if t, err := template.New(TypeToString(typ)).Parse(query); err == nil {
		var buf bytes.Buffer
		input := qryData{
			Depth:       depth,
			Name:        name,
			SubBalance:  newTxnValue("ROUND(SUM(CAST(s.value_num AS DOUBLE) / CAST (s.value_denom AS DOUBLE)), 2)", "", time1, time2),
			SubCredits:  newTxnValue("ROUND(SUM(CAST(s.value_num AS DOUBLE) / CAST (s.value_denom AS DOUBLE)), 2)", "AND s.value_num < 0", time1, time2),
			SubDebits:   newTxnValue("ROUND(SUM(CAST(s.value_num AS DOUBLE) / CAST (s.value_denom AS DOUBLE)), 2)", "AND s.value_num > 0", time1, time2),
			SubTxnCount: newTxnValue("COUNT(*)", "", time1, time2),
			Type:        TypeToString(typ),
			TypeValues:  TypeToValues(typ),
			Time1:       time1.Unix(),
			Time2:       time2.Unix(),
		}
		if err = t.Execute(&buf, input); err == nil {
			qry = buf.String()
		}
	}
	return qry
}

func StringToType(str string) (Type, error) {
	str = strings.ToUpper(str)
	for i, v := range typeStrings {
		if str == v {
			return Type(i), nil
		}
	}
	return All, syscall.EINVAL
}

func TypeToString(typ Type) string {
	return typeStrings[typ]
}

func TypeToValues(typ Type) string {
	vals := ""
	switch typ {
	case Asset:
		vals = fmt.Sprintf("'%s', '%s', '%s'", TypeToString(Asset), TypeToString(Bank), TypeToString(Cash))
	case Expense:
		vals = fmt.Sprintf("'%s'", TypeToString(Expense))
	case Income:
		vals = fmt.Sprintf("'%s'", TypeToString(Income))
	case Liability:
		vals = fmt.Sprintf("'%s', '%s'", TypeToString(Credit), TypeToString(Liability))
	default:
	}
	return vals
}
