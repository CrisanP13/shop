build:
	go build -o bin/main src/main.go

run:
	go run src/main.go

test: build
	go test -v ./test

test-full: migrate-down migrate clean test

.PHONY: test

migrate:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" up

migrate-down:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" down -all

clean:
	rm -r bin
