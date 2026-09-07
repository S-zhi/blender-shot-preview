KITEX ?= kitex
IDL := idl/v0_1/shot_preview.thrift
LLM_IDL := idl/v0_1/llm_gateway.thrift

.PHONY: generate test run

generate:
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(IDL)
	$(KITEX) -module github.com/S-zhi/blender-shot-preview $(LLM_IDL)

test:
	go test ./...

run:
	go run ./cmd/server
