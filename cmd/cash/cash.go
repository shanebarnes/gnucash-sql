package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanebarnes/gnucash-sql/internal/account"
)

func writeReport(writer io.Writer, db *sqlx.DB, query string) (int, error) {
	records := 0
	rows, err := db.Queryx(query)
	if err == nil {
		defer rows.Close()
		for err == nil && rows.Next() {
			result := account.Row{}
			err = rows.StructScan(&result)
			if err == nil {
				records++
				gocsv.MarshalWithoutHeaders([]*account.Row{&result}, writer)
			}
		}
	}

	return records, err
}

func rangeWriteReport(writer io.Writer, db *sqlx.DB, typ account.Type, maxDepth int, t1, t2 time.Time) {
	var err error
	for depth := 1; depth <= maxDepth && err == nil; depth++ {
		if _, err = writeReport(writer, db, account.NewQuery(typ, depth, t1, t2)); err != nil {
			fmt.Fprintln(writer, "Failed to execute query:", err)
		}
	}
}

func main() {
	dbFile := flag.String("db", "", "GnuCash SQLite3 database file name")
	depth := flag.Int("depth", 2, "expense account depth")
	year := flag.Int("year", 2020, "expense year")
	flag.Parse()

	writer := os.Stdout
	if db, err := sqlx.Connect("sqlite3", *dbFile); err == nil {
		defer db.Close()
		gocsv.Marshal([]*account.Row{}, writer)

		t0 := time.Time{}
		t1, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-01-01T00:00:00+00:00", *year))
		t2, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-12-31T23:59:59+00:00", *year))

		rangeWriteReport(writer, db, account.Asset, *depth, t0, t2)
		rangeWriteReport(writer, db, account.Expense, *depth, t1, t2)
		rangeWriteReport(writer, db, account.Income, *depth, t1, t2)
		rangeWriteReport(writer, db, account.Liability, *depth, t0, t2)
	} else {
		fmt.Fprintln(writer, "Failed to open database:", err)
	}
}
