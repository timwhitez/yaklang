#!/bin/sh
go build -gcflags=-l -ldflags="-s -w"  -o ~/Downloads/main  ./main.go  