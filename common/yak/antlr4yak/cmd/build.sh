#!/bin/sh
go test . 
GOOS=linux GOARCH=amd64 go build -gcflags=-l -ldflags="-s -w" -o ~/Downloads/yakvm_geek2023/main  ./main.go
cp ./marshalCode ~/Downloads/yakvm_geek2023/marshalCode
