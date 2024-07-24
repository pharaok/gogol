BINARY_NAME=gogol.wasm

build:
	tinygo build -target wasm -o ./assets/${BINARY_NAME} ./cmd/wasm 
