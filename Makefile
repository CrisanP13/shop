build:
	go build -o bin/main src/main.go

run:
	go run src/main.go

test: migrate-drop migrate
	mysql -uroot -pqwer shop < ./test/demo-data.sql
	go test -v ./test

.PHONY: test

migrate:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" up

migrate-drop:
	migrate -source file://src/migrations/ -database "mysql://root:qwer@tcp(127.0.0.1:3306)/shop" drop -f

clean:
	rm -r bin
