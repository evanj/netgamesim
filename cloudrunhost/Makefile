BUILD_OUTPUT_DIR:=build

all:
	mkdir -p $(BUILD_OUTPUT_DIR)
	GOOS=js GOARCH=wasm go build -o $(BUILD_OUTPUT_DIR)/wasmanim.wasm ../wasmanim
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" $(BUILD_OUTPUT_DIR)
	cp ../wasmanim/*.html $(BUILD_OUTPUT_DIR)
	docker build . "--tag=us.gcr.io/evan-jones-hrd/gamesim:$(shell date +%Y%m%d)-$(shell git rev-parse --short HEAD)"

push: all
	docker push "us.gcr.io/evan-jones-hrd/gamesim:$(shell date +%Y%m%d)-$(shell git rev-parse --short HEAD)"

clean:
	$(RM) -r $(BUILD_OUTPUT_DIR)
