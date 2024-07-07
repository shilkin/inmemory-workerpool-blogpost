start:
	docker-compose up -d --remove-orphans
.PHONY: start

stop:
	docker-compose down --remove-orphans
.PHONY: stop

refresh:
	docker-compose up -d --build
.PHONY: refresh

load:
	docker-compose up -d loadgen --build
.PHONY: load

restart: stop start
.PHONY: restart