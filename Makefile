run:
	docker-compose up --build
down:
	docker-compose down
tidy:
	go mod tidy
generate:
	oapi-codegen -generate gin,types,strict-server -o ./generated/api.gen.go ./schema.yaml