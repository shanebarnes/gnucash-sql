#!/bin/bash

go vet -v ./...
go test -v ./... -cover
go build -v -o bin/report cmd/report/report.go
