package account

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTime1() time.Time {
	return time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func TestNewQuery_Asset(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Asset, 1, t1, t1.Add(time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('ASSET', 'BANK', 'CASH')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836801\"\n"))
}

func TestNewQuery_Expense(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Expense, 2, t1, t1.Add(2 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('EXPENSE')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836802\"\n"))
}

func TestNewQuery_Income(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Income, 3, t1, t1.Add(3 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('INCOME')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836803\"\n"))
}

func TestNewQuery_Liability(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Liability, 4, t1, t1.Add(4 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('CREDIT', 'LIABILITY')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836804\"\n"))
}

func TestTypeToString(t *testing.T) {
	assert.Equal(t, "'ASSET', 'BANK', 'CASH'", TypeToString(Asset))
	assert.Equal(t, "'EXPENSE'", TypeToString(Expense))
	assert.Equal(t, "'INCOME'", TypeToString(Income))
	assert.Equal(t, "'CREDIT', 'LIABILITY'", TypeToString(Liability))
}
