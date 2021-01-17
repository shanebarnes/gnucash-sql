# gnucash-sql
SQLite Queries for use with GnuCash

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
```
