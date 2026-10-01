SHELL := bash

.PHONY: init-env check-env-acr sync-configs login-acr pull-images
.PHONY: up down logs ps restart down-v
.PHONY: up-images prod-deploy prod-pull prod-up prod-logs prod-ps prod-restart prod-down

COMPOSE := docker compose -f docker-compose.yml
COMPOSE_ACR := docker compose -f docker-compose.yml -f docker-compose.acr.yml
APP_IMAGES := arguscms argussms argusweb

init-env:
	@test -f .env || cp .env.example .env
	@echo "已准备 .env，请检查 ADVERTISE_IP、WEB_PORT 与 ALIYUN_ACR_*"

check-env-acr:
	chmod +x scripts/check-env-acr.sh
	./scripts/check-env-acr.sh

sync-configs:
	chmod +x scripts/sync-configs.sh
	./scripts/sync-configs.sh

login-acr:
	@set -a && source <(sed 's/\r$$//' .env) && set +a && \
	test -n "$$ALIYUN_ACR_REGISTRY" || (echo "请在 .env 配置 ALIYUN_ACR_REGISTRY"; exit 1) && \
	echo "$$ALIYUN_ACR_PASSWORD" | docker login "$$ALIYUN_ACR_REGISTRY" -u "$$ALIYUN_ACR_USER" --password-stdin

pull-images: check-env-acr login-acr
	@set -a && source <(sed 's/\r$$//' .env) && set +a && \
	for img in $(APP_IMAGES); do \
	  echo ">>> pull $${ACR_IMAGE_PREFIX}/$$img:$${IMAGE_TAG:-latest}"; \
	  docker pull "$${ACR_IMAGE_PREFIX}/$$img:$${IMAGE_TAG:-latest}"; \
	done

prod-pull: pull-images

up: init-env sync-configs
	$(COMPOSE) up -d --build

up-images: init-env sync-configs pull-images
	$(COMPOSE_ACR) up -d --no-build --remove-orphans

prod-up: up-images

prod-deploy:
	chmod +x scripts/prod-deploy.sh
	./scripts/prod-deploy.sh

down:
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

restart:
	$(COMPOSE) restart

prod-logs:
	$(COMPOSE_ACR) logs -f

prod-ps:
	$(COMPOSE_ACR) ps

prod-restart:
	$(COMPOSE_ACR) restart

prod-down:
	$(COMPOSE_ACR) down
