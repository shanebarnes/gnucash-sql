@echo off

go env || exit /b
gofmt -w . || exit /b
go vet -v ./... || exit /b
go test -v ./... -cover || exit /b
go build -v -o bin\report-windows.exe cmd\report\report.go || exit /b
