build: 
	go build .

build-cli:
	go build -tags cli .

dev:
	make build && ./jaqen

dev-cli:
	make build-cli && ./jaqen

cli:
	go run -tags cli .