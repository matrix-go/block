build:
	@go build -p=1 -o bin/block

run: build
	@./bin/block
	
test:
	@go test -p=1 -v ./...