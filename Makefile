OUT=bin/release
ENTRY=cmd/main.go

build:
	go build -o $(OUT) $(ENTRY)

run:
	@go run $(ENTRY)

.PHONY: build run
