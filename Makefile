
BINDIR=$(CURDIR)/bin
SERVER_IMAGE=revredir-server
CLIENT_IMAGE=revredir-client

run-server:
	go run github.com/pav5000/reverse-redirector/cmd/server

run-client:
	go run github.com/pav5000/reverse-redirector/cmd/client

local-bin:
	mkdir -p ${BINDIR}

build-server: local-bin
	go build -o ${BINDIR}/server github.com/pav5000/reverse-redirector/cmd/server

build-client: local-bin
	go build -o ${BINDIR}/client github.com/pav5000/reverse-redirector/cmd/client

docker-server:
	sudo docker build --build-arg APP=server -t ${SERVER_IMAGE} .

docker-client:
	sudo docker build --build-arg APP=client -t ${CLIENT_IMAGE} .

start-server:
	sudo docker run -d \
		--name=${SERVER_IMAGE} \
		--restart=always \
		--network=host \
		-v `pwd`/server.yml:/server.yml \
		${SERVER_IMAGE}

stop-server:
	sudo docker stop ${SERVER_IMAGE}; true
	sudo docker rm -f ${SERVER_IMAGE}; true

restart-server: stop-server start-server
	echo "restarted server"

logs-server:
	sudo docker logs --tail 100 -f ${SERVER_IMAGE}

start-client:
	sudo docker run -d \
		--name=${CLIENT_IMAGE} \
		--restart=always \
		--network=host \
		-v `pwd`/client.yml:/client.yml \
		${CLIENT_IMAGE}

stop-client:
	sudo docker stop ${CLIENT_IMAGE}; true
	sudo docker rm -f ${CLIENT_IMAGE}; true

restart-client: stop-client start-client
	echo "restarted client"

logs-client:
	sudo docker logs --tail 100 -f ${CLIENT_IMAGE}
