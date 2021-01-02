package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanebarnes/gnucash-sql/internal/account/expense"
)

func writeExpenses(writer io.Writer, db *sqlx.DB, depth, year int) (int, error) {
	records := 0
	rows, err := db.Queryx(expense.Query, fmt.Sprintf("%d-", year), depth)
	if err == nil {
		defer rows.Close()
		for err == nil && rows.Next() {
			result := expense.Row{}
			err = rows.StructScan(&result)
			if err == nil {
				records++
				gocsv.MarshalWithoutHeaders([]*expense.Row{&result}, writer)
			}
		}
	}

	return records, err
}

func main() {
	dbFile := flag.String("db", "", "GnuCash SQLite3 database file name")
	depth := flag.Int("depth", 2, "expense account depth")
	year := flag.Int("year", 2020, "expense year")
	flag.Parse()

	writer := os.Stdout
	if db, err := sqlx.Connect("sqlite3", *dbFile); err == nil {
		defer db.Close()
		gocsv.Marshal([]*expense.Row{}, writer)
		for i := 1; i <= *depth && err == nil; i++ {
			if _, err = writeExpenses(writer, db, i, *year); err != nil {
				fmt.Fprintln(writer, "Failed to execute expense query:", err)
			}
		}
	} else {
		fmt.Fprintln(writer, "Failed to open database:", err)
	}
}
