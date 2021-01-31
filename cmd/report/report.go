package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/jinzhu/now"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shanebarnes/gnucash-sql/internal/account"
)

const (
	flagAccount = "account"
	flagName = "name"
	flagDatabase = "db"
	flagDepth = "depth"
	flagDateEnd = "end"
	//flagDateGroupBy = "groupby"
	flagDateMonthOf = "monthof"
	flagDateQuarterOf = "quarterof"
	flagDateStart = "start"
	flagDateWeekOf = "weekof"
	flagDateYearOf = "yearof"
)

type args struct {
	account      string
	accountType  account.Type
	dbFile       string
	date         string
	depth        int
	dtEnd        time.Time
	dtStart      time.Time
	name         string
	strEnd       string
	strGroupBy   string
	strMonthOf   string
	strQuarterOf string
	strStart     string
	strWeekOf    string
	strYearOf    string
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

func rangeWriteReport(writer gocsv.CSVWriter, db *sqlx.DB, typ account.Type, name string, maxDepth int, t1, t2 time.Time) {
	var err error
	for depth := 1; depth <= maxDepth && err == nil; depth++ {
		if _, err = writeReport(writer, db, account.NewQuery(typ, name, depth, t1, t2)); err != nil {
			fmt.Println("Failed to execute query:", err)
		}
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&conf.account, flagAccount, "all", "account type (all, asset, income, expense, liability)")
	flag.StringVar(&conf.name, flagName, account.QueryWildcard, "account name")
	flag.StringVar(&conf.dbFile, flagDatabase, "", "GnuCash SQLite3 database file name")
	flag.IntVar(&conf.depth, flagDepth, 2, "account report depth")
	flag.StringVar(&conf.strEnd, flagDateEnd, now.With(time.Now()).EndOfYear().String(), "report end date")
	flag.StringVar(&conf.strStart, flagDateStart, now.With(time.Now()).BeginningOfYear().String(), "report start date")
	//flag.StringVar(&conf.strGroupBy, flagDateGroupBy, "", "report group interval (week, month, quarter, year)")
	flag.StringVar(&conf.strMonthOf, flagDateMonthOf, now.With(time.Now()).BeginningOfMonth().String(), "report month of date")
	flag.StringVar(&conf.strQuarterOf, flagDateQuarterOf, now.With(time.Now()).BeginningOfQuarter().String(), "report quarter of date")
	flag.StringVar(&conf.strWeekOf, flagDateWeekOf, now.With(time.Now()).BeginningOfWeek().String(), "report week of date")
	flag.StringVar(&conf.strYearOf, flagDateYearOf, now.With(time.Now()).BeginningOfYear().String(), "report year of date")
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	exitOnError(isArgsValid(&conf))
}

func isArgsValid(a *args) error {
	var err error

	flagsDateStartEnd := isFlagPassed(flagDateStart, flagDateEnd)
	flagsDateOf := isFlagPassed(flagDateMonthOf, flagDateQuarterOf, flagDateWeekOf, flagDateYearOf)

	if a.accountType, err = account.StringToType(a.account); err != nil {
		err = fmt.Errorf("Invalid account: %v", err)
	} else if _, err = os.Stat(a.dbFile); err != nil {
		err = fmt.Errorf("Invalid database file name: %v", err)
	} else if conf.depth < 1 {
		err = fmt.Errorf("Account report depth must be greater than zero")
	} else if flagsDateOf != 0 && flagsDateOf != 1 {
		err = fmt.Errorf("Too many report dates provided")
	} else if flagsDateStartEnd != 0 && flagsDateOf != 0 {
		err = fmt.Errorf("Too many report dates provided")
	} else if flagsDateStartEnd > 0 || (flagsDateStartEnd == 0 && flagsDateOf == 0) {
		if conf.dtStart, err = now.Parse(conf.strStart); err != nil {
			err = fmt.Errorf("Invalid report start date: %v", err)
		} else if conf.dtEnd, err = now.Parse(conf.strEnd); err != nil {
			err = fmt.Errorf("Invalid report end date: %v", err)
		}
		// TODO: apply group by flag?
	} else if flagsDateOf > 0 {
		if isFlagPassed(flagDateMonthOf) > 0 {
			if conf.dtStart, err = now.Parse(conf.strMonthOf); err == nil {
				conf.dtStart = now.With(conf.dtStart).BeginningOfMonth()
				conf.dtEnd = now.With(conf.dtStart).EndOfMonth()
			} else {
				err = fmt.Errorf("Invalid report month of date: %v", err)
			}
		} else if isFlagPassed(flagDateQuarterOf) > 0 {
			if conf.dtStart, err = now.Parse(conf.strQuarterOf); err == nil {
				conf.dtStart = now.With(conf.dtStart).BeginningOfQuarter()
				conf.dtEnd = now.With(conf.dtStart).EndOfQuarter()
			} else {
				err = fmt.Errorf("Invalid report quarter of date: %v", err)
			}
		} else if isFlagPassed(flagDateWeekOf) > 0 {
			if conf.dtStart, err = now.Parse(conf.strWeekOf); err == nil {
				conf.dtStart = now.With(conf.dtStart).BeginningOfWeek()
				conf.dtEnd = now.With(conf.dtStart).EndOfWeek()
			} else {
				err = fmt.Errorf("Invalid report week of date: %v", err)
			}
		} else if isFlagPassed(flagDateYearOf) > 0 {
			if conf.dtStart, err = now.Parse(conf.strYearOf); err == nil {
				conf.dtStart = now.With(conf.dtStart).BeginningOfYear()
				conf.dtEnd = now.With(conf.dtStart).EndOfYear()
			} else {
				err = fmt.Errorf("Invalid report year of date: %v", err)
			}
		}
	}

	if isFlagPassed(flagName) > 0 {
		if !strings.HasPrefix(conf.name, account.QueryWildcard) {
			conf.name = account.QueryWildcard + conf.name
		}
		if !strings.HasSuffix(conf.name, account.QueryWildcard) {
			conf.name = conf.name + account.QueryWildcard
		}
	}

	return err
}

func isFlagPassed(names ...string) int {
	count := 0
	for _, name := range names {
		flag.Visit(func(f *flag.Flag) {
			if f.Name == name {
				count++
			}
		})
	}
	return count
}

func main() {
	if db, err := sqlx.Connect("sqlite3", conf.dbFile); err == nil {
		defer db.Close()

		writer := gocsv.DefaultCSVWriter(os.Stdout)
		//writer.Comma = '\t'
		gocsv.MarshalCSV([]*account.Row{}, writer)

		t0 := time.Time{}
		t1 := conf.dtStart
		t2 := conf.dtEnd

		switch conf.accountType {
		case account.All:
			rangeWriteReport(writer, db, account.Asset, conf.name, conf.depth, t0, t2)
			rangeWriteReport(writer, db, account.Expense, conf.name, conf.depth, t1, t2)
			rangeWriteReport(writer, db, account.Income, conf.name, conf.depth, t1, t2)
			rangeWriteReport(writer, db, account.Liability, conf.name, conf.depth, t0, t2)
		case account.Asset, account.Bank, account.Cash, account.Credit, account.Liability:
			rangeWriteReport(writer, db, conf.accountType, conf.name, conf.depth, t0, t2)
		default:
			rangeWriteReport(writer, db, conf.accountType, conf.name, conf.depth, t1, t2)
		}
	} else {
		exitOnError(fmt.Errorf("Failed to open database: %v", err))
	}
}
