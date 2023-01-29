DB_DSN := postgres://admin:admin@192.168.49.2:30010/auth?sslmode=disable

migratecreate:
	migrate create -ext sql -dir db/migration -seq ${f}

migrateup:
	migrate -path db/migration -database "${DB_DSN}" -verbose up ${v}

migratedown:
	migrate -path db/migration -database "${DB_DSN}" -verbose down ${v}

migrateforce:
	migrate -path db/migration -database "${DB_DSN}" -verbose force ${v}

protogen:
	protoc --proto_path=proto proto/auth_service.proto proto/user_service.proto proto/general.proto \
	--go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative

.PHONY: sqlcgen
sqlcgen:
	sqlc generate

.PHONY: rebuild
rebuild:
	docker build -t ngoctd/ecommerce-auth:latest . && \
	docker push ngoctd/ecommerce-auth

.PHONY: redeploy
redeploy:
	kubectl rollout restart deployment depl-auth

.PHONY: migratecreate migrateup migratedown migrateforce protogen