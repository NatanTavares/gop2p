IP ?= 127.0.0.1
PORT ?= 54468
PEER_ID ?= 12D3KooWH1fLw4arQ8cuTUAMettfz6roC4nRusBsdgmsTdDsMa33

.PHONY run:
run:
	@go run main.go

.PHONY run-client:
run-client:
	@go run main.go --peer-address /ip4/$(IP)/tcp/$(PORT)/p2p/$(PEER_ID)