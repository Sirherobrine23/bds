DB_NAME=bds
PASSWORD=GoBDds3433@

postgres_docker_image=postgres:16
mysql_docker_image=mysql:9
mssql_docker_image=mcr.microsoft.com/mssql/server:2022-latest

all: build

build:
	go generate
	go build -o go_bds -v .

run:
	go generate
	ENCRYPT_PASSWORD=$(PASSWORD) go run -v . web

postgres:
	docker pull $(postgres_docker_image)
	docker run -e "POSTGRES_DB=$(DB_NAME)" \
		-e "POSTGRES_USER=postgres" \
		-e "POSTGRES_PASSWORD=$(PASSWORD)" \
		--network some-network -d $(postgres_docker_image)

mysql:
	docker pull $(mysql_docker_image)
	docker run -e "MYSQL_DATABASE=$(DB_NAME)" \
		-e "MYSQL_ROOT_PASSWORD=$(PASSWORD)" \
		--network some-network -d $(mysql_docker_image)

mssql:
	docker pull $(mssql_docker_image)
	docker run -e "ACCEPT_EULA=Y" \
		-e "MSSQL_SA_PASSWORD=$(PASSWORD)" \
		--network some-network \
		--name sql1 --hostname sql1 -d $(mssql_docker_image)