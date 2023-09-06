all: help

APP=$(shell basename $(shell pwd))
PROTOBUF=proto/

# default golang docker image
GOLANG_IMG=minixxie/golang:1.21.0
# default migrate docker image
MIGRATE_IMG=migrate/migrate:4
# default mariadb docker image
MARIADB_IMG=mariadb:10.3
MYSQL_ROOT_PASSWORD=hello123

JSON_HEADER="Content-type: application/json; charset=utf-8"

ANSI_GREEN="\x1B[32m"
ANSI_RESET="\x1B[0m"

SHELL := /bin/bash

#####   ####   ####  #    # ###### #####
#    # #    # #    # #   #  #      #    #
#    # #    # #      ####   #####  #    #
#    # #    # #      #  #   #      #####
#    # #    # #    # #   #  #      #   #
#####   ####   ####  #    # ###### #    #

help:
	###########################################################################################################
	# [DOCKER]
	# make local_network        - create docker bridge network "local_network"
	# make ps                   - docker ps -a (list all containers)
	# make up                   - run docker-compose up (run up the container)
	# make down                 - run docker-compose down (shutdown the container)
	# make kill                 - run docker-compose rm -f (kill and rm the container)
	# make restart              - run docker-compose restart (restart the container)
	# make logs                 - tail the container logs
	# make clean                - run: make rm-con, make rm-img
	# make initdb               - initialize any DBs on local (current project need to provide ./db/ and ./db/migrations/)
	# make conndb-root          - connect to MariaDB using root
	# make conndb               - connect to MariaDB using DB-user
	# make conndb-app           - connect to MariaDB using APP-user
	# make connredis3           - connect to Redis 3.2
	# make connredis4           - connect to Redis 4.0
	# make stats                - show container stats (CPU%, memory, etc)
	# make stats-all            - show all containers stats (CPU%, memory, etc)
	# make sh                   - enter the container
	# make cli                  - run a new container as client on HOST network (golang docker image)
	# make ubuntu-cli           - run a new container as client on local_network network (ubuntu docker image)
	# make golang               - run a new container with code mounted (golang docker image)
	# rm-con                    - remove all dead containers (non-zero Exited)
	# rm-img                    - remove all <none> images/layers
	#
	# [DOCKER PUBLIC]
	# nginx                     - run up an nginx-proxy container
	# nginx-down                - remove the nginx-proxy container
	# redis3                    - run up an redis container (v3.2)
	# redis3-down               - remove the redis3 container
	# redis4                    - run up an redis container (v4.0)
	# redis3-down               - remove the redis4 container
	# mongo                     - run up an mongo container
	# mongo-down                - remove the mongo container
	# mariadb10.0               - run up an mariadb container (v10.0)
	# mariadb10.0-down          - remove the mariadb container (v10.0)
	# mariadb10.3               - run up an mariadb container (v10.3)
	# mariadb10.3-down          - remove the mariadb container (v10.3)
	# postgis                   - run up an postgis container (v10)
	# postgis-down              - remove the postgis container
	# postgres                  - run up an postgres container (v10)
	# postgres-down             - remove the postgres container
	# tidb                      - run up an tidb container
	# tidb-down                 - remove the tidb container
	#
	# [PROJECT]
	# init                      - initialize the project with necessary folder structure
	#
	# [MICROSERVICE]
	# protoc                    - compile *.proto, to golang code into src/vendor/
	# gotest                    - run "go test" in container
	# gofmt                     - format golang source code (change the codes)
	# godoc                     - serve godoc for source code and open in browser (Mac)
	# redoc                     - serve redoc for gRPC gateway and open in browser (Mac)
	# curl                      - passing variables, it starts a container and do curl/ab on your endpoint
	###########################################################################################################
	@echo "Enjoy!"

local_network:
	docker network create -d bridge local_network || true

ps:
	docker ps -a

up: local_network
	DOCKER_BUILDKIT=0 docker-compose -f docker-compose.local.yml up --build -d

down:
	docker-compose -f docker-compose.local.yml down

kill:
	docker-compose -f docker-compose.yml kill && \
	docker-compose -f docker-compose.local.yml rm -f

restart:
	docker-compose -f docker-compose.local.yml restart

logs:
	docker-compose -f docker-compose.local.yml logs -f --tail=10

initdb:
	@if [ -f "$$PWD"/initdb.sh ]; then \
		"$$PWD"/initdb.sh; \
	elif [ -f "$$PWD"/db/Makefile ]; then \
		cd "$$PWD"/db; make; \
	else \
		if ! [ -d "./db/" ]; then echo "ERR: ./db/ doesn't exist"; exit 1; fi; \
		if ! [ -f "./db/db.rc" ]; then echo "ERR: ./db/db.rc doesn't exist"; exit 1; fi; \
		if ! [ -d "./db/migrations/" ]; then echo "ERR: ./db/migrations/ doesn't exist"; exit 1; fi; \
		source "./db/db.rc"; echo "DB: $$DB"; \
		rootPass=${MYSQL_ROOT_PASSWORD}; \
		a=$(eval TMP := $(shell mktemp -d ~/XXXXXX)); trap 'rm -rf $(TMP)' EXIT; \
		echo ""; echo "1) Drop & Create DB..."; \
		echo "DROP DATABASE IF EXISTS \`$$DB\`;" > $(TMP)/db.sql; \
		echo "CREATE DATABASE \`$$DB\` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci;" >> $(TMP)/db.sql; \
		if [ "$$mariadb" == "mariadb10.3" ]; then \
			echo "DROP USER IF EXISTS '$$DB'@'%';" >> $(TMP)/db.sql; \
		fi; \
		echo "CREATE USER '$$DB'@'%' IDENTIFIED BY '$$DB';" >> $(TMP)/db.sql; \
		echo "GRANT ALL ON \`$$DB\`.* TO '$$DB'@'%';" >> $(TMP)/db.sql; \
		echo "FLUSH PRIVILEGES;" >> $(TMP)/db.sql; \
		docker run -it --rm --network=local_network -v $(TMP)/db.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
		echo "DB user was created to manage schema change, you can connect with: make conndb"; \
		echo ""; echo "2) Create schema..."; \
		docker run -it --rm --network=local_network -v "${PWD}"/db/migrations:/migrations "${MIGRATE_IMG}" -path /migrations -database "mysql://$$DB:$$DB@tcp($$mariadb:3306)/$$DB" up; \
		echo ""; echo "3) Create App user(s)..."; \
		for u in $$APP_USERS; do \
			if [ "$$mariadb" == "mariadb10.3" ]; then \
				echo "DROP USER IF EXISTS '$$u'@'%';" > $(TMP)/user-$$u.sql; \
			elif [ "$$mariadb" == "mariadb10.0" ]; then \
				echo "DROP USER '$$u'@'%';" > $(TMP)/user-$$u.sql; \
			fi; \
			echo "CREATE USER '$$u'@'%' IDENTIFIED BY '$$u';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT,INSERT,UPDATE,DELETE ON \`$$DB\`.* TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT ON \`$$DB\`.\`schema_migrations\` TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "FLUSH PRIVILEGES;" >> $(TMP)/user-$$u.sql; \
			docker run -it --rm --network=local_network -v $(TMP)/user-$$u.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
		done; \
		echo "App users were created to simulate the connection from the app, you can connect with: make conndb-app"; \
		echo ""; echo "4) Running customized SQL files..."; \
		for sql in `ls ./db/*.sql | sort` ; do \
			echo "$$sql..."; \
			time docker run -it --rm --network=local_network -v "${PWD}/$$sql":/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass $$DB < /file.sql" ; \
			echo "$$sql...[OK]"; \
		done; \
	fi

conndb-root:
	@if ! [ -d "./db/" ]; then echo "ERR: ./db/ doesn't exist"; exit 1; fi; \
	if [ -f "./db/db.rc" ]; then source "./db/db.rc"; fi; \
	docker run -it --rm --network=local_network ${MARIADB_IMG} bash -c "mysql -A --default-character-set=utf8 -h$$mariadb -uroot -p"

conndb:
	@if ! [ -d "./db/" ]; then echo "ERR: ./db/ doesn't exist"; exit 1; fi; \
	if [ -f "./db/db.rc" ]; then source "./db/db.rc"; fi; \
	docker run -it --rm --network=local_network ${MARIADB_IMG} bash -c "mysql -A --default-character-set=utf8 -h$$mariadb -u$$DB -p$$DB $$DB"

conndb-app:
	@if ! [ -d "./db/" ]; then echo "ERR: ./db/ doesn't exist"; exit 1; fi; \
	if [ -f "./db/db.rc" ]; then source "./db/db.rc"; fi; \
	if [ "$${#APP_USERS[@]}" -gt 1 ]; then \
		i=0; \
		while [ "$${APP_USERS[i]}" != "" ]; do \
			echo "$$i) "$${APP_USERS[i]}; \
			i=$$(($$i + 1)); \
		done; \
		count=$$(($$i - 1)); \
		echo -n "Choose (0-$count) to connect mysql: "; read index; \
		u=$${APP_USERS[$$index]}; \
	else \
		u=$${APP_USERS[0]}; \
	fi; \
	docker run -it --rm --network=local_network ${MARIADB_IMG} bash -c "mysql -A --default-character-set=utf8 -h$$mariadb -u$$u -p$$u $$DB"

connredis3:
	docker run --rm -it --net=local_network redis:3.2-alpine redis-cli -h redis3

connredis4:
	docker run --rm -it --net=local_network redis:4.0-alpine redis-cli -h redis4


clean: rm-img

stats:
	docker stats local_${APP}

stats-all:
	docker stats `docker ps -a | sed 1d | awk '{print $$NF}'`

sh:
	docker-compose -f docker-compose.local.yml exec local_${APP} bash

cli:
	docker run --rm -it --net=host "${GOLANG_IMG}" bash

ubuntu-cli: local_network
	docker run --rm -it --net=local_network ubuntu:18.04 bash

rm-con:
	deads=$$(docker ps -a | sed 1d | grep "Exited " | grep -v "Exited (0)" | awk '{print $$1}'); if [ "$$deads" != "" ]; then docker rm -f $$deads; fi

rm-img: rm-con
	none=$$(docker images | sed 1d | grep "^<none>" | awk '{print $$3}'); if [ "$$none" != "" ]; then docker rmi $$none; fi

#####   ####   ####  #    # ###### #####     #####  #    # #####  #      #  ####
#    # #    # #    # #   #  #      #    #    #    # #    # #    # #      # #    #
#    # #    # #      ####   #####  #    #    #    # #    # #####  #      # #
#    # #    # #      #  #   #      #####     #####  #    # #    # #      # #
#    # #    # #    # #   #  #      #   #     #      #    # #    # #      # #    #
#####   ####   ####  #    # ###### #    #    #       ####  #####  ###### #  ####

nginx: local_network
	docker rm -f nginx-proxy || true
	docker run -d --name=nginx-proxy --hostname=nginx-proxy --restart=always \
		-p 80:80 -p 443:443 --net=local_network \
		-v /var/run/docker.sock:/tmp/docker.sock:ro \
		jwilder/nginx-proxy:alpine

nginx-down:
	docker rm -f nginx-proxy || true

redis3: local_network
	# add "-p 6379:7379" if you want to publish the port
	docker rm -f redis3 || true
	docker run -d --name=redis3 --hostname=redis3 --restart=always \
		--net=local_network \
		redis:3.2-alpine
redis4: local_network
	# add "-p 6379:7379" if you want to publish the port
	docker rm -f redis4 || true
	docker run -d --name=redis4 --hostname=redis4 --restart=always \
		--net=local_network \
		redis:4.0-alpine

redis3-down:
	docker rm -f redis3 || true

redis4-down:
	docker rm -f redis4 || true

# mongo:
# 	docker rm -f mongo || true
# 	docker run -d --name=mongo --hostname=mongo --restart=always \
# 		-p 27017:27017 --net=local_network \
# 		mongo:3.7-jessie \
# 		mongod --replSet "rs0" --auth
# 	sleep 30s ; docker exec -i mongo mongo --eval "rs.initiate();"
# 	sleep 10s ; if [ "$$MONGO_ROOT_PASSWORD" == "" ] ; then MONGO_ROOT_PASSWORD=hello123 ; fi ; \
# 		docker exec -i mongo mongo admin --eval \
# 			'use admin; db.createUser({user:"siteUserAdmin",pwd:"hello123",roles[{role:"userAdminAnyDatabase",db:"admin"}]});'
# 	# sleep 10s ; if [ "$$MONGO_ROOT_PASSWORD" == "" ] ; then MONGO_ROOT_PASSWORD=hello123 ; fi ; \
# 	# 	docker exec -i mongo mongo admin --eval \
# 	# 		"use admin; db.createUser({user:\"root\",pwd:\"root\",roles:[\"root\"]});"

mongo-down:
	docker rm -f mongo || true

mariadb10.0: local_network
	# add "-p 3306:3306" if you want to publish the port
	docker rm -f mariadb || true; \
	docker run -d --name=mariadb10.0 --hostname=mariadb10.0 --restart=always \
		--net=local_network \
		-e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
		-e TERM=dumb \
		mariadb:10.0 \
		mysqld --lower-case-table-names=1

mariadb10.0-down:
	docker rm -f mariadb10.0 || true

mariadb10.3: local_network
	# add "-p 3306:3306" if you want to publish the port
	docker rm -f mariadb10.3 || true; \
	docker run -d --name=mariadb10.3 --hostname=mariadb10.3 --restart=always \
		--net=local_network \
		-e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
		-e TERM=dumb \
		mariadb:10.3 \
		mysqld --lower-case-table-names=1

mariadb10.3-down:
	docker rm -f mariadb10.3 || true

postgis: local_network
	docker rm -f postgis || true
	if [ "$$POSTGRES_PASSWORD" == "" ] ; then POSTGRES_PASSWORD=hello123 ; fi ; \
	docker run -d --name=postgis --hostname=postgis --restart=always \
		-p 5432:5432 --net=local_network \
		-e POSTGRES_PASSWORD=$$POSTGRES_PASSWORD \
		mdillon/postgis:10-alpine

postgis-down:
	docker rm -f postgis || true

postgres9.6: local_network
	docker rm -f postgres9.6 || true
	if [ "$$POSTGRES_PASSWORD" == "" ] ; then POSTGRES_PASSWORD=hello123 ; fi ; \
	docker run -d --name=postgres9.6 --hostname=postgres9.6 --restart=always \
		-p 5432:5432 --net=local_network \
		-e POSTGRES_PASSWORD=$$POSTGRES_PASSWORD \
		postgres:9.6-alpine

postgres9.6-down:
	docker rm -f postgres9.6 || true

postgres: local_network
	docker rm -f postgres || true
	if [ "$$POSTGRES_PASSWORD" == "" ] ; then POSTGRES_PASSWORD=hello123 ; fi ; \
	docker run -d --name=postgres --hostname=postgres --restart=always \
		-p 5432:5432 --net=local_network \
		-e POSTGRES_PASSWORD=$$POSTGRES_PASSWORD \
		postgres:10-alpine

postgres-down:
	docker rm -f postgres || true

tidb: local_network
	docker rm -f tidb || true
	docker run -d --name=tidb --hostname=tidb --restart=always \
		-p 4000:4000 --net=local_network \
		pingcap/tidb:v1.0.8

tidb-down:
	docker rm -f tidb || true

###### #    #  ####  #    # #####  ######
#      ##   # #      #    # #    # #
#####  # #  #  ####  #    # #    # #####
#      #  # #      # #    # #####  #
#      #   ## #    # #    # #   #  #
###### #    #  ####   ####  #    # ######

init:
	mkdir -p src/"${PROTOBUF}" src/services/ src/vendor/
	echo "vendor/*" > .gitignore
	echo "package main" > src/main.go


#    # #  ####  #####   ####      ####  ###### #####  #    # #  ####  ######
##  ## # #    # #    # #    #    #      #      #    # #    # # #    # #
# ## # # #      #    # #    #     ####  #####  #    # #    # # #      #####
#    # # #      #####  #    #         # #      #####  #    # # #      #
#    # # #    # #   #  #    #    #    # #      #   #   #  #  # #    # #
#    # #  ####  #    #  ####      ####  ###### #    #   ##   #  ####  ######


golang: local_network
	docker run --rm -it --net=local_network -v "${PWD}/config/local.yml:/config/local.yml" \
		-v "${PWD}/../..:/micro" \
		-v "${PWD}:/go/src/app" \
		-w "/go/src/app" -e ENV=local "${GOLANG_IMG}" bash

gotest: local_network protoc
	docker run --rm -it --net=host -v "${PWD}:/go/src/app" -w "/go/src/app" -e ENV=local "${GOLANG_IMG}" bash -c "go test ./..."

gofmt:
	docker run --rm -t -v "${PWD}:/go/src/app" -w "/go/src/app" "${GOLANG_IMG}" gofmt -w .

godoc:
	@docker rm -f local_$(APP)_godoc || true
	@while [ true ]; do \
		PORT=$$(( ( RANDOM % 60000 )  + 1025 )); \
		nc -z -vv localhost $$PORT >/dev/null 2>/dev/null || break; \
	done; \
	docker run --rm -it --net=local_network -v "${PWD}:/go/src/app" -w "/go/src/app" \
		--expose=80 -p "$$PORT:80" \
		-e ENV=local -d --name=local_$(APP)_godoc "${GOLANG_IMG}" bash -c \
		"godoc -http :80";\
	docUrl="http://localhost:$$PORT/pkg/$(APP)"; \
	sleep 3s ; os=$$(uname -s); if [ "$$os" == Darwin ]; then echo "Opening browser..."; open "$$docUrl"; fi; \


redoc: protoc
	@docker rm -f local_$(APP)_redoc || true
	@while [ true ]; do \
		PORT=$$(( ( RANDOM % 60000 )  + 1025 )); \
		nc -z -vv localhost $$PORT >/dev/null 2>/dev/null || break; \
	done; \
	if [ -f ".redoc" ]; then redoc=$$(cat .redoc); fi; \
	docker run -it --net=local_network \
		-d --name=local_$(APP)_redoc \
		-v "${PWD}:/go/src/app" -w "/go/src/app" \
		--expose=80 \
		-p "$$PORT:80" \
		-e VIRTUAL_HOST=redoc.$(APP).local.test.com \
		-e VIRTUAL_PORT=80 \
		-e REDOC=$$redoc \
		"${GOLANG_IMG}" bash -c \
		'if [ "$$REDOC" == "" ]; then swg=$$(find ./vendor -name "*swagger.json" | head -n 1); else swg=$$(find ./vendor -name "*"$$REDOC"*swagger.json" | head -n 1); fi; echo swg=$$swg; swagger serve --no-open --base-path=/ -F redoc -p $$VIRTUAL_PORT $$swg'; \
	docUrl=http://localhost:$$PORT/docs; \
	echo ""; echo "Access Redoc: " ; \
	echo "1) http://localhost:$$PORT/docs ; OR" ; \
	echo "2) http://redoc.$(APP).local.test.com/docs (adding \"127.0.0.1  redoc.$(APP).local.test.com\" into /etc/hosts)" ; echo ""; \
	sleep 3s ; os=$$(uname -s); if [ "$$os" == Darwin ]; then echo "Opening browser..."; open "$$docUrl"; fi; \
	docker logs -f --tail=10 local_$(APP)_redoc ; echo $$?

protoc:
	@echo "Under construction"
