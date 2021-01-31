package account

import (
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTime1() time.Time {
	return time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func TestNewQuery_Asset(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Asset, "%", 1, t1, t1.Add(time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "SELECT 'ASSET' AS account_type,\n"))
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('ASSET', 'BANK', 'CASH')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836801\""))
}

func TestNewQuery_Expense(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Expense, "%", 2, t1, t1.Add(2 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "SELECT 'EXPENSE' AS account_type,\n"))
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('EXPENSE')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836802\""))
}

func TestNewQuery_Income(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Income, "%", 3, t1, t1.Add(3 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "SELECT 'INCOME' AS account_type,\n"))
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('INCOME')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836803\""))
}

func TestNewQuery_Liability(t *testing.T) {
	t1 := getTime1()
	qry := NewQuery(Liability, "%", 4, t1, t1.Add(4 * time.Second))
	assert.NotEmpty(t, qry)
	assert.True(t, strings.Contains(qry, "SELECT 'LIABILITY' AS account_type,\n"))
	assert.True(t, strings.Contains(qry, "AND tr.account_type IN ('CREDIT', 'LIABILITY')\n"))
	assert.True(t, strings.Contains(qry, "AND strftime('%s', tx.post_date) BETWEEN \"1577836800\" AND \"1577836804\""))
}

func TestStringToType(t *testing.T) {
	typ, err := StringToType("")
	assert.Equal(t, All, typ)
	assert.Equal(t, syscall.EINVAL, err)

	typ, err = StringToType("all")
	assert.Equal(t, All, typ)
	assert.Nil(t, err)

	typ, err = StringToType("ASSET")
	assert.Equal(t, Asset, typ)
	assert.Nil(t, err)

	typ, err = StringToType("Bank")
	assert.Equal(t, Bank, typ)
	assert.Nil(t, err)

	typ, err = StringToType("CaSh")
	assert.Equal(t, Cash, typ)
	assert.Nil(t, err)

	typ, err = StringToType("credit")
	assert.Equal(t, Credit, typ)
	assert.Nil(t, err)

	typ, err = StringToType("expense")
	assert.Equal(t, Expense, typ)
	assert.Nil(t, err)

	typ, err = StringToType("income")
	assert.Equal(t, Income, typ)
	assert.Nil(t, err)

	typ, err = StringToType("liability")
	assert.Equal(t, Liability, typ)
	assert.Nil(t, err)
}

func TestTypeToString(t *testing.T) {
	assert.Equal(t, "ALL", TypeToString(All))
	assert.Equal(t, "ASSET", TypeToString(Asset))
	assert.Equal(t, "BANK", TypeToString(Bank))
	assert.Equal(t, "CASH", TypeToString(Cash))
	assert.Equal(t, "CREDIT", TypeToString(Credit))
	assert.Equal(t, "EXPENSE", TypeToString(Expense))
	assert.Equal(t, "INCOME", TypeToString(Income))
	assert.Equal(t, "LIABILITY", TypeToString(Liability))
}

func TestTypeToValues(t *testing.T) {
	assert.Equal(t, "'ASSET', 'BANK', 'CASH'", TypeToValues(Asset))
	assert.Equal(t, "'EXPENSE'", TypeToValues(Expense))
	assert.Equal(t, "'INCOME'", TypeToValues(Income))
	assert.Equal(t, "'CREDIT', 'LIABILITY'", TypeToValues(Liability))
}
