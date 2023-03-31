up:	
	docker-compose -f "docker-compose.yaml" up -d --build
down: 
	docker-compose -f "docker-compose.yaml" down
crnet:
	docker network create shoplist
rebuild:
	docker-compose build --no-cache
remdb:
	rm ./db/shoplist.db
abot:
	docker exec -it shoplist_bot /bin/sh -c "[ -e /bin/bash ] && /bin/bash || /bin/sh"
ashop:
	docker exec -it shoplist_server /bin/sh -c "[ -e /bin/bash ] && /bin/bash || /bin/sh"
gen:
	ent generate ./internal/shoplist/ent/schema
deploy:
	git pull
	sudo systemctl stop shoplist
	sudo cp ./shoplist.service /etc/systemd/system/shoplist.service
	systemctl enable shoplist
	systemctl start shoplist
	systemctl -l status shoplist