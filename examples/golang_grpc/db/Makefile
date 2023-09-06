# default migrate docker image
MIGRATE_IMG=migrate/migrate:4
# default mariadb docker image
MARIADB_IMG=mariadb:10.3
MYSQL_ROOT_PASSWORD=hello123

SHELL := /bin/bash

all:
	@if ! [ -d "./migrations/" ]; then echo "ERR: ./migrations/ doesn't exist"; exit 1; fi; \
	if ! [ -f "./db.rc" ]; then echo "ERR: ./db.rc doesn't exist"; exit 1; fi; \
	source "./db.rc"; echo "DB: $$DB"; \
	rootPass=${MYSQL_ROOT_PASSWORD}; \
	a=$(eval TMP := $(shell mktemp -d ~/XXXXXX)); trap 'rm -rf $(TMP)' EXIT; \
	echo ""; echo "1) Drop & Create DB..."; \
	if [ "$$mariadb" == "mariadb10.3" ]; then \
		echo "DROP DATABASE IF EXISTS \`$$DB\`;" > $(TMP)/db.sql; \
		echo "CREATE DATABASE \`$$DB\` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci;" >> $(TMP)/db.sql; \
		echo "DROP USER IF EXISTS '$$DB'@'%';" >> $(TMP)/db.sql; \
		echo "CREATE USER '$$DB'@'%' IDENTIFIED BY '$$DB';" >> $(TMP)/db.sql; \
		echo "GRANT ALL ON \`$$DB\`.* TO '$$DB'@'%';" >> $(TMP)/db.sql; \
		echo "FLUSH PRIVILEGES;" >> $(TMP)/db.sql; \
		docker run -it --rm --network=local_network -v $(TMP)/db.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
	elif [ "$$mariadb" == "mariadb10.0" ]; then \
		echo "DROP DATABASE IF EXISTS \`$$DB\`;" > $(TMP)/db.sql; \
		echo "CREATE DATABASE \`$$DB\` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci;" >> $(TMP)/db.sql; \
		echo "CREATE USER '$$DB'@'%' IDENTIFIED BY '$$DB';" >> $(TMP)/db.sql; \
		echo "GRANT ALL ON \`$$DB\`.* TO '$$DB'@'%';" >> $(TMP)/db.sql; \
		echo "FLUSH PRIVILEGES;" >> $(TMP)/db.sql; \
		docker run -it --rm --network=local_network -v $(TMP)/db.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
	fi; \
	echo "DB user was created to manage schema change, you can connect with: make conndb"; \
	echo ""; echo "2) Create schema..."; \
	docker run -it --rm --network=local_network -v "${PWD}"/migrations:/migrations "${MIGRATE_IMG}" -path /migrations -database "mysql://$$DB:$$DB@tcp($$mariadb:3306)/$$DB" up; \
	echo ""; echo "3) Create App user(s)..."; \
	for u in $$APP_USERS; do \
		if [ "$$mariadb" == "mariadb10.3" ]; then \
			echo "DROP USER IF EXISTS '$$u'@'%';" > $(TMP)/user-$$u.sql; \
			echo "CREATE USER '$$u'@'%' IDENTIFIED BY '$$u';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT,INSERT,UPDATE,DELETE ON \`$$DB\`.* TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT ON \`$$DB\`.\`schema_migrations\` TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "FLUSH PRIVILEGES;" >> $(TMP)/user-$$u.sql; \
			docker run -it --rm --network=local_network -v $(TMP)/user-$$u.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
		elif [ "$$mariadb" == "mariadb10.0" ]; then \
			echo "DROP USER '$$u'@'%';" > $(TMP)/user-$$u.sql; \
			echo "CREATE USER '$$u'@'%' IDENTIFIED BY '$$u';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT,INSERT,UPDATE,DELETE ON \`$$DB\`.* TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "GRANT SELECT ON \`$$DB\`.\`schema_migrations\` TO '$$u'@'%';" >> $(TMP)/user-$$u.sql; \
			echo "FLUSH PRIVILEGES;" >> $(TMP)/user-$$u.sql; \
			docker run -it --rm --network=local_network -v $(TMP)/user-$$u.sql:/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass < /file.sql"; \
		fi; \
	done; \
	echo "App users were created to simulate the connection from the app, you can connect with: make conndb-app"; \
	echo ""; echo "4) Running customized SQL files..."; \
	for sql in `ls *.sql | sort` ; do \
		echo "$$sql..."; \
		time docker run -it --rm --network=local_network -v "${PWD}/$$sql":/file.sql ${MARIADB_IMG} bash -c "mysql -A -h$$mariadb -uroot -p$$rootPass $$DB < /file.sql" ; \
		echo "$$sql...[OK]"; \
	done