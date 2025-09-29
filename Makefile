include .env
export $(shell sed -n 's/^\([^#]*\)=.*/\1/p' .env)

APP := bin/main

build:
	@go build -o $(APP) src/main.go

run:
	@go run src/main.go

test: e2e-test
	@echo Testing...

E2E_PACKAGES :=	 ./test
e2e-test: migrate-drop migrate add-demo-data build
	@echo "Starting app..."
	@bin/main > e2e.log 2>&1 & echo $$! > app.pid
	@sleep 2 #wait for startup

	@EXIT_CODE=0; \
	for pckg in $(E2E_PACKAGES); do \
		echo "Running tests in $$pckg"; \
		go test -v $$pckg || EXIT_CODE=$$?; \
	done; \
	@echo "Stopping app..."; \
	kill `cat app.pid`; \
	wait `cat app.pid` 2>/dev/null || true; \
	rm -f app.pid; \
	exit $$EXIT_CODE

migrate:
	@migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" up

migrate-drop:
	@migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" drop -f

add-demo-data:
	@mysql -uroot -pqwer shop < ./test/demo-data.sql

clean:
	@rm -r ./bin

# set-env
# set -a && . ./.env && set +a

# clear-env
# unset $(grep -v '^#' .env | sed -E 's/(.*)=.*/\1/' | xargs)
