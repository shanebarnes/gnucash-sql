# gnucash-sql

Run SQL queries quickly and effortlessly against your GnuCash SQLite database

## Build Instructions

### Unix

```
$ git clone https://github.com/shanebarnes/gnucash-sql.git
$ cd gnucash-sql
$ ./build/build.sh
```

### Windows

```
> git clone https://github.com/shanebarnes/gnucash-sql.git
> cd gnucash-sql
> scripts\build.cmd

```

## Examples

```
# Run with defaults (get all accounts report for the current year)
./bin/report -db sqlite3.gnucash

# Get yearly expense report
./bin/report -db sqlite3.gnucash -account expense -yearof 2021

# Get quarterly income report
./bin/report -db sqlite3.gnucash -account income -quarterof 2021-1

# Get monthly accounts report
./bin/report -db sqlite3.gnucash -monthof 2021-1

# Get weekly accounts report
./bin/report -db sqlite3.gnucash -weekof 2021-1

# Get accounts report for a specific period of time
./bin/report -db sqlite3.gnucash -start 2021-1-2 -end 2021-1-16

# Get accounts report for since a specific start date
./bin/report -db sqlite3.gnucash -start 2021-1-2

# Get accounts report by searching for a specific name
./bin/report -db sqlite3.gnucash -yearof 2020 -name utilities

# Get accounts report for the last decade grouped by year
./bin/report -db sqlite3.gnucash -start 2013 -groupby year
```
