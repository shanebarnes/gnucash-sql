package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/jinzhu/now"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanebarnes/gnucash-sql/internal/account"
)

const (
	dateFormat = "2006-01-02"
)

type args struct {
	dbFile    string
	date      string
	depth     int
	dt        time.Time
	monthly   bool
	quarterly bool
	weekly    bool
	yearly    bool
}
var conf args

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

func init() {
	flag.StringVar(&conf.dbFile, "db", "", "GnuCash SQLite3 database file name")
	flag.StringVar(&conf.date, "date", time.Now().Format(dateFormat), "expense date")
	flag.IntVar(&conf.depth, "depth", 2, "expense account depth")
	flag.BoolVar(&conf.yearly, "yearly", false, "yearly report")
	flag.BoolVar(&conf.quarterly, "quarterly", false, "quarterly report")
	flag.BoolVar(&conf.monthly, "monthly", false, "monthly report")
	flag.BoolVar(&conf.weekly, "weekly", false, "weekly report")
	flag.Parse()

	var err error
	if conf.dt, err = time.Parse(dateFormat, conf.date); err != nil {
		panic("Invalid date")
	} else if !conf.monthly && !conf.quarterly && !conf.weekly && !conf.yearly {
		panic("No reporting period specified")
	} else if conf.monthly && conf.quarterly && conf.weekly && conf.yearly {
		panic("More than one reporting period specified")
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	if db, err := sqlx.Connect("sqlite3", conf.dbFile); err == nil {
		defer db.Close()

		loc := time.Now().Location()
		t0 := time.Time{}

		ts := time.Date(conf.dt.Year(), conf.dt.Month(), conf.dt.Day(), 0, 0, 0, 0, loc)
		var t1, t2 time.Time
		if conf.monthly {
			t1 = now.With(ts).BeginningOfMonth()
			t2 = now.With(ts).EndOfMonth()
		} else if conf.quarterly {
			t1 = now.With(ts).BeginningOfQuarter()
			t2 = now.With(ts).EndOfQuarter()
		} else if conf.weekly {
			t1 = now.With(ts).BeginningOfWeek()
			t2 = now.With(ts).EndOfWeek()
		} else if conf.yearly {
			t1 = now.With(ts).BeginningOfYear()
			t2 = now.With(ts).EndOfYear()
		}

		writer := gocsv.DefaultCSVWriter(os.Stdout)
		//writer.Comma = '\t'

		gocsv.MarshalCSV([]*account.Row{}, writer)
		rangeWriteReport(writer, db, account.Asset, conf.depth, t0, t2)
		rangeWriteReport(writer, db, account.Expense, conf.depth, t1, t2)
		rangeWriteReport(writer, db, account.Income, conf.depth, t1, t2)
		rangeWriteReport(writer, db, account.Liability, conf.depth, t0, t2)
	} else {
		fmt.Println("Failed to open database:", err)
	}
}
