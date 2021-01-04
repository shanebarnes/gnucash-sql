package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanebarnes/gnucash-sql/internal/account"
)

func writeReport(writer gocsv.CSVWriter, db *sqlx.DB, query string) (int, error) {
	records := 0
	rows, err := db.Queryx(query)
	if err == nil {
		defer rows.Close()
		for err == nil && rows.Next() {
			result := account.Row{}
			err = rows.StructScan(&result)
			if err == nil {
				records++
				gocsv.MarshalCSVWithoutHeaders([]*account.Row{&result}, writer)
			}
		}
	}

	return records, err
}

func rangeWriteReport(writer gocsv.CSVWriter, db *sqlx.DB, typ account.Type, maxDepth int, t1, t2 time.Time) {
	var err error
	for depth := 1; depth <= maxDepth && err == nil; depth++ {
		if _, err = writeReport(writer, db, account.NewQuery(typ, depth, t1, t2)); err != nil {
			fmt.Println("Failed to execute query:", err)
		}
	}
}

func main() {
	dbFile := flag.String("db", "", "GnuCash SQLite3 database file name")
	depth := flag.Int("depth", 2, "expense account depth")
	year := flag.Int("year", 2020, "expense year")
	flag.Parse()

	if db, err := sqlx.Connect("sqlite3", *dbFile); err == nil {
		defer db.Close()

		loc := time.Now().Location()
		t0 := time.Time{}
		t1 := time.Date(*year, 1, 1, 0, 0, 0, 0, loc)
		t2 := time.Date(*year, 12, 31, 23, 59, 59, 999999999, loc)

		writer := gocsv.DefaultCSVWriter(os.Stdout)
		//writer.Comma = '\t'

		gocsv.MarshalCSV([]*account.Row{}, writer)
		rangeWriteReport(writer, db, account.Asset, *depth, t0, t2)
		rangeWriteReport(writer, db, account.Expense, *depth, t1, t2)
		rangeWriteReport(writer, db, account.Income, *depth, t1, t2)
		rangeWriteReport(writer, db, account.Liability, *depth, t0, t2)
	} else {
		fmt.Println("Failed to open database:", err)
	}
}
