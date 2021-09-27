MIGRATIONS_FOLDER:="assets/migrations"
PROTOBUF_FOLDER:="internal/satellite_communication/*.proto"
USER:= root
PASSWORD:= emis
MYSQLHOST:=localhost
PORT:=3306
DBNAME:=satellites
GOOS:=windows
PRODUCT:= satelliteApp
REPO:=github.com/Simek13
GOARCH:=amd64
PROTO_GOOGLE_CMD=$(shell go list -f '{{ .Dir }}' -mod=mod -m github.com/grpc-ecosystem/grpc-gateway)/third_party/googleapis

ifeq ($(GOOS),windows)
PROTO_GOOGLE=$(subst \,/,$(PROTO_GOOGLE_CMD))
else
PROTO_GOOGLE=$(PROTO_GOOGLE_CMD)
endif

################################################################################
# BUILD
################################################################################

.PHONY: build-deps
build-deps:
	go mod tidy

.PHONY: build
build: build-deps
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(PRODUCT) "$(REPO)/$(PRODUCT)/cmd/$(PRODUCT)"

################################################################################
# DATABASE
################################################################################

.PHONY: db-deps
db-deps: export GO111MODULE=off
db-deps:
	go get -u -d github.com/golang-migrate/migrate/cli
	go get -u github.com/go-sql-driver/mysql
	go build -tags 'mysql' -o $(GOPATH)/bin/migrate github.com/golang-migrate/migrate/cli

.PHONY: db-up
db-up: MIGRATE_BIN:="migrate"
db-up: db-deps ## run a database upgrade
	$(MIGRATE_BIN) -path $(MIGRATIONS_FOLDER) -database \
	mysql://$(USER):$(PASSWORD)@tcp\($(MYSQLHOST):$(PORT)\)/$(DBNAME) up

.PHONY: db-down
db-down: MIGRATE_BIN:="migrate"
db-down: db-deps ## run a database downgrade
	$(MIGRATE_BIN) -path $(MIGRATIONS_FOLDER) -database \
	mysql://$(USER):$(PASSWORD)@tcp\($(MYSQLHOST):$(PORT)\)/$(DBNAME) down

################################################################################
# TESTING
################################################################################

.PHONY: lint-deps
lint-deps: ## get linter for testing
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: lint
lint: lint-deps ## get linter for testing
	staticcheck ./...

################################################################################
# PROTOBUF
################################################################################

.PHONY: pb-gen
pb-gen: PROTOBUF_BIN:="protoc"
pb-gen: 
	$(PROTOBUF_BIN) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=$(PROTOBUF_FOLDER)

.PHONY: proto-deps
proto-deps: ## get the depedencies for generating gRPC libraries
#go: module github.com/golang/protobuf is deprecated: Use the "google.golang.org/protobuf" module instead.
	go get -d google.golang.org/protobuf/proto
	go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go get -d github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go get -d google.golang.org/protobuf/cmd/protoc-gen-go
	go get -d google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: proto-go
proto: proto-deps ## generate go gRPC libraries from the proto files
	protoc -I $(PROTO_GOOGLE)   \
	--go_out=./pkg --go_opt=paths=source_relative --go-grpc_out=./pkg --go-grpc_opt=paths=source_relative --proto_path=api/protobuf-spec satellite_communication.proto

	protoc -I $(PROTO_GOOGLE)  \
	--grpc-gateway_out=./pkg --grpc-gateway_opt=logtostderr=true --grpc-gateway_opt=paths=source_relative --proto_path=api/protobuf-spec satellite_communication.proto

	protoc -I $(PROTO_GOOGLE)  \
	--openapiv2_out ./pkg --openapiv2_opt logtostderr=true --proto_path=api/protobuf-spec satellite_communication.proto
