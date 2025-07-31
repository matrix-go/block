build:
	@go build -o bin/block

run: build
	@./bin/block
	
test:
	@go test -v ./...