include scripts/go.mk

BIN_DIR      := $(shell readlink -f ./bin)
BUILD_DIR    := $(shell readlink -f ./.build)
CACHE_DIR    := $(shell readlink -f ./.cache)
CERTS_DIR    := $(shell readlink -f ./cfssl/certs)
RESOURCE_DIR := $(shell readlink -f ./resources)
# Ensure the dirs above exist on a clean checkout
$(shell mkdir -p $(BIN_DIR) $(BUILD_DIR) $(CACHE_DIR) $(CERTS_DIR) $(RESOURCE_DIR))

# Swagger version to package and deploy
SWAGGER_UI_VERSION := 2.2.8
protoc := ./bin/protoc
protoc += --plugin=protoc-gen-go=$(GOPATH)/bin/protoc-gen-go
protoc += --proto_path=$(shell go list -f '{{ .Dir }}' -m github.com/gogo/protobuf)/protobuf:.
protoc += --proto_path=$(shell go list -f '{{ .Dir }}' -m github.com/golang/protobuf)/protoc-gen-go:.
protoc += --proto_path=$(shell go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway)/third_party/googleapis:.
protoc += -I .

.PHONY: all
all: deps check revendor code_gen resource_gen cert_gen compile  ## run all targets

.PHONY: gosources
gosources:
	@echo $(GOSOURCES) | sed -e 's/\s\+/\n/g'

.PHONY: deps
deps: _deps  ## install host dependencies
	@# install the arm cross compiler
	@if ! which arm-linux-gnueabihf-gcc-5 > /dev/null; then \
	  sudo apt-get -yq install gcc-5-arm-linux-gnueabihf; \
	fi
	@# install sqlite3 cli client
	@if ! which sqlite3 > /dev/null; then \
	  sudo apt-get -yq install sqlite3; \
	fi
	@# install ttyrec
	@if ! which ttyrec > /dev/null; then \
	  sudo apt-get -yq install ttyrec; \
	fi
	./scripts/get-protoc bin/protoc

.PHONY: check
check: _check  ## checks

.PHONY: revendor
revendor: check go.mod  ## install/build all 3rd party vendor libs and bins
	@go mod tidy -v
	@go mod vendor -v
	@go mod verify
	go install -v ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install -v ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go install -v ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go install -v ./vendor/github.com/cloudflare/cfssl/cmd/cfssl
	go install -v ./vendor/github.com/cloudflare/cfssl/cmd/cfssljson

.PHONY: code_gen
code_gen: check code_gen_helper  ## generate grpc go files from proto spec
	@# Helper only exists to make the default target doc work with line continuation

.PHONY: code_gen_helper
code_gen_helper: \
	pb/app.pb.go pb/app.pb.gw.go pb/app.swagger.json \
	example/pb/app.pb.go example/pb/app.pb.gw.go example/pb/app.swagger.json

pb/app.pb.go: pb/app.proto
	@# Generate the GRPC definitons from the .proto file
	$(protoc) \
	  --go_out=plugins=grpc:. \
	  pb/app.proto

pb/app.pb.gw.go: pb/app.proto
	@# Generate the GRPC Gateway which proxies to JSON from the .proto file
	$(protoc) \
	  --grpc-gateway_out=logtostderr=true:. \
	  --plugin=protoc-gen-grpc-gateway=$(GOPATH)/bin/protoc-gen-grpc-gateway \
	  pb/app.proto

pb/app.swagger.json: pb/app.proto
	@# Generate the swagger definition from the .proto file
	$(protoc) \
	  --swagger_out=logtostderr=true:. \
	  pb/app.proto

example/pb/app.pb.go: example/pb/app.proto
	@# Generate the GRPC definitons from the .proto file
	$(protoc) \
	  --go_out=plugins=grpc:. \
	  example/pb/app.proto

example/pb/app.pb.gw.go: example/pb/app.proto
	@# Generate the GRPC Gateway which proxies to JSON from the .proto file
	$(protoc) \
	  --grpc-gateway_out=logtostderr=true:. \
	  example/pb/app.proto

example/pb/app.swagger.json: example/pb/app.proto
	@# Generate the swagger definition from the .proto file
	$(protoc) \
	  --swagger_out=logtostderr=true:. \
	  example/pb/app.proto

.PHONY: resource_gen
resource_gen: check $(RESOURCE_DIR)/swagger/ui/ui.go $(RESOURCE_DIR)/swagger/files/files.go  ## generate go-bindata swagger files

$(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION):
	@# Download and extract the swagger-ui release
	export SWAGGER_UI_TGZ=$(CACHE_DIR)/swagger-ui-$(SWAGGER_UI_VERSION).tar.gz; \
	if [ ! -f $$SWAGGER_UI_TGZ ]; then \
	  curl -fsSL https://github.com/swagger-api/swagger-ui/archive/v$(SWAGGER_UI_VERSION).tar.gz > \
	    $$SWAGGER_UI_TGZ; \
	fi; \
	if [ ! -d $(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION) ]; then \
	  tar xzf $$SWAGGER_UI_TGZ -C $(BUILD_DIR); \
	fi
	@# Set the creation date to current, which marks regeneration of ui.go and files.go
	touch $(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION)

$(RESOURCE_DIR)/swagger/ui/ui.go: $(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION)
	@# Generate the swagger-ui directory as a golang file
	@# Ignore the warning about "Cannot read bindata.go open bindata.go: no such file or directory"
	mkdir -p $(RESOURCE_DIR)/swagger/ui
	cd $(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION)/dist && \
	  go-bindata-assetfs -o $(RESOURCE_DIR)/swagger/ui/ui.go -pkg ui 2>/dev/null ./... || true

$(RESOURCE_DIR)/swagger/files/files.go: $(BUILD_DIR)/swagger-ui-$(SWAGGER_UI_VERSION)
	@# Generate the swagger.json file as a golang file
	mkdir -p $(RESOURCE_DIR)/swagger/files
	mkdir -p $(BUILD_DIR)/swagger/files
	cp -f $(CURDIR)/pb/app.swagger.json $(BUILD_DIR)/swagger/files/swagger.json
	cd $(BUILD_DIR)/swagger/files && \
	  go-bindata-assetfs -o $(RESOURCE_DIR)/swagger/files/files.go -pkg files 2>/dev/null ./... || true

.PHONY: cert_gen
cert_gen: $(RESOURCE_DIR)/certs/certs.go  ## generate go-bindata cert files

cfssl/certs/insecure-key.pem:
	@# Call the Makefile in the subdir
	make -C cfssl

$(RESOURCE_DIR)/certs/certs.go: cfssl/certs/insecure-key.pem
	@# Generate the certs directory as a golang file
	@# Ignore the warning about "Cannot read bindata.go open bindata.go: no such file or directory"
	mkdir -p $(RESOURCE_DIR)/certs
	cd $(CERTS_DIR) && \
	  go-bindata-assetfs -o $(RESOURCE_DIR)/certs/certs.go -pkg certs ./... 2>/dev/null || true

.PHONY: compile
compile: check format $(BIN_DIR)/linux_amd64/$(BIN_NAME) ## build the binaries for amd64

.PHONY: compilex
compilex: check format $(BIN_DIR)/linux_amd64/$(BIN_NAME) $(BIN_DIR)/linux_arm/$(BIN_NAME)  ## build the binaries for all platforms

$(BIN_DIR)/linux_amd64/$(BIN_NAME): check $(GOSOURCES)
	@echo "## Building AMD64 Binary"
	export GOOS=linux; \
	export GOARCH=amd64; \
	go build $(GO_BUILD_FLAGS) \
	  -o "$(BIN_DIR)/$${GOOS}_$${GOARCH}/$(BIN_NAME)"

$(BIN_DIR)/linux_arm/$(BIN_NAME): check $(GOSOURCES)
	@echo "## Building ARM Binary"
	export GOOS=linux; \
	export GOARCH=arm; \
	export GOARM=7; \
	export CC=arm-linux-gnueabihf-gcc-5; \
	CGO_ENABLED=1 go build $(GO_BUILD_FLAGS) \
	  -o "$(BIN_DIR)/$${GOOS}_$${GOARCH}/$(BIN_NAME)"

.PHONY: format  ## run gofmt on all go sources
format: $(GOSOURCES) imports
	gofmt -w $(GOSOURCES)

.PHONY: imports  ## run goimports on all go sources
imports: $(GOSOURCES)
	goimports -w $(GOSOURCES)

.PHONY: test
test: _test format
	ginkgo -v -cover $(shell go list ./...)

.PHONY: testrandom
testrandom: _test format
	ginkgo -v -cover --randomizeSuites --randomizeAllSpecs ./app/... ./cmd/... ./resources/...

.PHONY: clean
clean:  ## delete all non-repo files
	rm -rf bin .ginkgo .build *.bak
	find ./ -type f -name '*.coverprofile' | xargs rm -f

.PHONY: demo
demo: ## run and record the demo-magic script
	ttyrec -e './demo/demo.sh' ./demo/recording.ttyrec
	ttyrec2gif -in ./demo/recording.ttyrec -out demo/demo.gif -s 1.0 -col 80 -row 24
	rm -f ./demo/recording.ttyrec
	rm -rf bin/

.PHONY: notes
notes:
	@### Notes. The following can be used to build a lib file
	@# cd ./app && \
	@#   go install -pkgdir="$(PKG_DIR)/$${GOOS}_$${GOARCH}"
	@# gcc-arm-linux-gnueabihf
	@# https://developer.ubuntu.com/en/snappy/guides/cross-build/
