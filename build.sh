#!/bin/bash

go vet -v ./...
go test -v ./... -cover
go build -v -o bin/cash cmd/cash/cash.go
