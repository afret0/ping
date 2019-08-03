#!/usr/bin/env bash
#rm goFrame_test
packr2
env GOOS=linux GOARCH=amd64 go build -o goFrame_test  main.go
scp goFrame_test test:/root/work/goFrame
packr2 clean
rm goFrame_test
