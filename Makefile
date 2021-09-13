MIGRATIONS_FOLDER:="assets/migrations"
USER:= root
PASSWORD:= emis
MYSQLHOST:=localhost
PORT:=3306
DBNAME:=satelites


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