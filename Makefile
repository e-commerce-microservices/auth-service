DB_DSN := postgres://admin:admin@192.168.49.2:30010/auth?sslmode=disable

migratecreate:
	migrate create -ext sql -dir db/migration -seq ${f}

migrateup:
	migrate -path db/migration -database "${DB_DSN}" -verbose up ${v}

migratedown:
	migrate -path db/migration -database "${DB_DSN}" -verbose down ${v}

migrateforce:
	migrate -path db/migration -database "${DB_DSN}" -verbose force ${v}

protogen_auth:
	protoc --proto_path=proto proto/auth/*.proto \
	--go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative

protogen_user:
	protoc --proto_path=proto proto/user/*.proto \
	--go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative


sqlcgen:
	sqlc generate

.PHONY: migratecreate migrateup migratedown migrateforce protogen_auth protogen_user