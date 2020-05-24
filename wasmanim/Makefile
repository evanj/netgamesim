all:
	GOOS=js GOARCH=wasm go vet .
	golint .
	GOOS=js GOARCH=wasm go build -o wasmanim.wasm .
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" .
	go run ../host
