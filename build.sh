#!/bin/bash

set -eu

gofmt -w .
go vet -v ./...
go test -v ./... -race -cover

out="./bin"
cmd="${out}/report-$(go env GOOS)$(go env GOEXE)"
cmdnative="${out}/report$(go env GOEXE)"

go build -v -o "$cmd" cmd/report/report.go
cp "$cmd" "$cmdnative"
