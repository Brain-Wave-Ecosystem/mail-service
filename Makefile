# Docker-Compose commands
mail-up:
	docker-compose -f ./deployments/compose/mail-service.docker-compose.yaml --env-file=./.env up -d --build

mail-down:
	docker-compose -f ./deployments/compose/mail-service.docker-compose.yaml --env-file=./.env down