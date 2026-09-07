KITEX ?= kitex
IDL := idl/v0_1/shot_preview.thrift

.PHONY: generate test run

generate:
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(IDL)

test:
	go test ./...

run:
	go run ./cmd/server
