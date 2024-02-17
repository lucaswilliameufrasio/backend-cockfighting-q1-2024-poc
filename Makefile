build-container:
	docker build --progress=plain -t gatling-on-docker -f Dockerfile .
.PHONY: build-container

start-container:
	docker run -it -d -v ./local-results:/opt/gatling/results --network=host --name gatling gatling-on-docker || echo 'Maybe it is already running'
.PHONY: start-container

run-simulation:
	docker exec gatling /opt/gatling/bin/gatling.sh -rm local -sf /opt/gatling/user-files/simulations -s RinhaBackendCrebitosSimulation -rf /opt/gatling/results
.PHONY: run-simulation
