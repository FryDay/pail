#!/bin/sh

~/go/bin/reflex -r '\.go' -s -- sh -c "go run cmd/pail/main.go"
