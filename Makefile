build:
	go build -o bin/main src/main.go

run:
	go run src/main.go

test: migrate-down migrate
	go test -v ./test

.PHONY: test

migrate:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" up

migrate-down:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" down -all

clean:
	rm -r bin
