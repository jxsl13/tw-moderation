default: start

start:
	docker-compose up -d
stop:
	docker-compose down

debug: start
	sleep 5
	docker logs subscriber