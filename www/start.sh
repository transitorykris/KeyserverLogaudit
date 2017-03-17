#!/bin/bash

go build

APP_DB="root:secret@tcp(localhost:3306)/logaudit?parseTime=true" \
./api
