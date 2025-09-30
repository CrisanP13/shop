
export $(shell sed -n 's/^\([^#]*\)=.*/\1/p' .env)

.PHONY: build run test e2e-test migrate migrate-drop add-demo-data clean

APP := bin/main

build:
	go build -o $(APP) src/main.go

run:
	go run src/main.go

test: e2e-test

E2E_PACKAGES :=	 ./test
e2e-test: migrate-drop migrate add-demo-data build
	@echo "Starting app"
	@( \
	  $(APP) > e2e.log 2>&1 & \
	  echo $$! > app.pid; \
	  trap "echo 'Stopping app...'; kill \`cat app.pid\`; wait \`cat app.pid\` 2>/dev/null || true; rm -f app.pid" EXIT; \
	  echo "Waiting for app"; \
	  retries=5; \
	  while [ $$retries -gt 0 ]; do \
	  if curl -s $(SHOP_ADDR):$(SHOP_PORT)/health > /dev/null; then \
		  echo "App started"; break; \
		fi; \
		retries=$$((retries-1)); \
		[ $$retries -eq 0 ] && echo "App failed to start" && exit 1; \
		sleep 0.5; \
	  done; \
	  EXIT_CODE=0; \
	  for pckg in $(E2E_PACKAGES); do \
		echo "Running tests in $$pckg"; \
		go test -v $$pckg || EXIT_CODE=$$?; \
	  done; \
	  exit $$EXIT_CODE; \
	)

migrate:
	@echo "Migrating"
	@migrate -source file://src/migrations/ -database "mysql://$$SHOP_DB_USER:$$SHOP_DB_PASS@tcp($$SHOP_DB_ADDR)/$$SHOP_DB_NAME" up

migrate-drop:
	@echo "Droping db"
	@migrate -source file://src/migrations/ -database "mysql://$$SHOP_DB_USER:$$SHOP_DB_PASS@tcp($$SHOP_DB_ADDR)/$$SHOP_DB_NAME" drop -f

add-demo-data:
	@echo "Adding demo data"
	@mysql -u$$SHOP_DB_USER -p$$SHOP_DB_PASS $$SHOP_DB_NAME < ./test/demo-data.sql

clean:
	rm -r ./bin

# set-env
# set -a && . ./.env && set +a

# clear-env
# unset $(grep -v '^#' .env | sed -E 's/(.*)=.*/\1/' | xargs)
