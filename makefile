DB_NAME=bds
PASSWORD=GoBDds3433@

POSTGRES_DOCKER_IMAGE=postgres:16
mysql_docker_image=mysql:9
mssql_docker_image=mcr.microsoft.com/mssql/server:2022-latest

# Test and build latest version
all: test build

# Generate JS files
gen:
	go generate

# Build latest bin file
build: gen
	go build -o go_bds -v .

# Run test
test: gen
	ENCRYPT_PASSWORD=$(PASSWORD) go test -v -timeout 0 ./...

# Start web
run: gen
	ENCRYPT_PASSWORD=$(PASSWORD) go run -v . web

# Start postgres server
postgres:
	docker pull $(POSTGRES_DOCKER_IMAGE)
	docker run -e "POSTGRES_DB=$(DB_NAME)" \
		-e "POSTGRES_USER=postgres" \
		-e "POSTGRES_PASSWORD=$(PASSWORD)" \
		--network some-network -d $(POSTGRES_DOCKER_IMAGE)

# Start mysql server
mysql:
	docker pull $(mysql_docker_image)
	docker run -e "MYSQL_DATABASE=$(DB_NAME)" \
		-e "MYSQL_ROOT_PASSWORD=$(PASSWORD)" \
		--network some-network -d $(mysql_docker_image)

# Start microsoft sql server
mssql:
	docker pull $(mssql_docker_image)
	docker run -e "ACCEPT_EULA=Y" \
		-e "MSSQL_SA_PASSWORD=$(PASSWORD)" \
		--network some-network \
		--name sql1 --hostname sql1 -d $(mssql_docker_image)