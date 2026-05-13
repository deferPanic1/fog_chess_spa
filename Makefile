include .env
export

DSN=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: run build tidy docker-up docker-down migrate-up migrate-down migrate-status

migrate-up:
	goose -dir migrations postgres "$(DSN)" up

migrate-down:
	goose -dir migrations postgres "$(DSN)" down

migrate-status:
	goose -dir migrations postgres "$(DSN)" status
