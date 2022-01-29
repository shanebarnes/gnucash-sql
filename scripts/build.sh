#!/bin/bash

set -eu

go env
gofmt -w .
go vet -v ./...
go test -v ./... -race -cover
go build -v -o "./bin/report-$(go env GOOS)$(go env GOEXE)" cmd/report/report.go
