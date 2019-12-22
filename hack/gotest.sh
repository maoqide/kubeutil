#! /bin/sh
# run unit test
source ./setenv.sh
go test $(go list ./...)