build-container:
	sudo docker build --progress=plain -t gatling-on-docker -f Dockerfile .
.PHONY: build-container

start-container:
	sudo docker start gatling
.PHONY: start-container

create-container:
	sudo docker run -it -d --network host -v ./local-results:/opt/gatling/results --name gatling gatling-on-docker || echo 'Maybe it is already running'
.PHONY: create-container

stop-container:
	sudo docker stop gatling && docker rm gatling
.PHONY: stop-container

run-simulation:
	sudo docker exec gatling /opt/gatling/bin/gatling.sh -rm local -sf /opt/gatling/user-files/simulations -s RinhaBackendCrebitosSimulation -rf /opt/gatling/results
.PHONY: run-simulation
