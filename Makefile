MIGRATIONS_FOLDER:="assets/migrations"
PROTOBUF_FOLDER:="internal/satellite_communication/satellite_communication.proto"
USER:= root
PASSWORD:= emis
MYSQLHOST:=localhost
PORT:=3306
DBNAME:=satellites
GOOS:=windows
PRODUCT:= satelliteApp
REPO:=github.com/Simek13
GOARCH:=amd64

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
	$(PROTOBUF_BIN) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTOBUF_FOLDER)
