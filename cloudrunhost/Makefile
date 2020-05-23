all:
	GOOS=js GOARCH=wasm go build -o wasmanim.wasm ../wasmanim
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" .
	cp ../wasmanim/index.html .
	docker build . "--tag=us.gcr.io/evan-jones-hrd/gamesim:$(shell date +%Y%m%d)-$(shell git rev-parse --short HEAD)"
