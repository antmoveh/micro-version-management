
SHELL := /bin/bash
BASEDIR = $(shell pwd)

# build with verison infos
buildDate = $(shell TZ=Asia/Shanghai date +%FT%T%z)
gitCommit = `git log --pretty='%s%b%B' -n 1`
app = app

build:
	echo $(gitCommit) -- $(buildDate)
	rm -f ${app}
	@go build -v -o $(app) .
help:
	@echo "make build - build app"

.PHONY: help

