SOURCES := $(wildcard *.go cmd/*/*.go pkg/*/*.go)

VERSION=$(shell git describe --tags --dirty 2>/dev/null)

ifeq ($(VERSION),)
	VERSION := "0.0.1-beta"
endif

.PHONY: build
build: $(SOURCES)
	make tidy
	mkdir -p bin/
	go build -o bin/cxi-k8s-device-plugin -ldflags "-X main.version=$(VERSION)" ./cmd
	@echo "Built. Version: $(VERSION)"

tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -rf bin/
	@echo "Cleaned build artifacts."

.PHONY: run	
run: clean 
	make build
	./bin/cxi-k8s-device-plugin -logtostderr=true -stderrthreshold=INFO -v=5

.PHONY: test
test: tidy
	@echo "Running tests..."
	@echo "Version: $(VERSION)"
	@echo "Running unit tests..."
	go test -v ./pkg/...