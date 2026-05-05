.PHONY: up down build logs clean

up:
	cd deployments && docker-compose up -d
	@echo "Все сервисы запущены"
	@echo "Go сервер: http://localhost:8081"
	@echo "Python API: http://localhost:8000"

down:
	cd deployments && docker-compose down

build:
	cd deployments && docker-compose build --no-cache

logs:
	cd deployments && docker-compose logs -f

clean:
	cd deployments && docker-compose down -v
	docker system prune -f

pull-model:
	cd deployments && docker-compose exec ollama ollama pull mistral

restart:
	cd deployments && docker-compose restart

status:
	cd deployments && docker-compose ps