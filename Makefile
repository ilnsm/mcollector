.PHONY: all agent server
all: server agent

agent:
	go build -o cmd/agent/agent ./cmd/agent/*.go
server:
	go build -o cmd/server/server ./cmd/server/*.go