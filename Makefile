default: start

start: build
	docker-compose up -d
stop:
	docker-compose down

build:
	docker-compose build --force-rm

debug: start
	sleep 5
	docker logs detect-vpn

logs:
	docker logs detect-vpn

clean:
	docker system prune -f
	-rm -f detect-vpn/detect-vpn
	-rm -f publisher/publisher
	-rm -f subscriber/subscriber