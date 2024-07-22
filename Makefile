BINARY_NAME=go.wasm

build:
	GOOS=js GOARCH=wasm go build -o ./assets/${BINARY_NAME} ./cmd/wasm/
