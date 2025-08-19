build:
	@go build -p=1 -o bin/block

run: build
	@./bin/block
	
test:
	@go test -p=1 --cover -v ./...

mint:
	@go test -p=1 -v -run=TestSendMintTx -tags=e2e .