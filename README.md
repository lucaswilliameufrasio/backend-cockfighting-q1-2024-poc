# Backend Cockfighting Q1 2024 POC

Note1: :warning: **This README does not cover the Windows and Mac process** :warning:
Note2: If you are a Docker Rootless user, i suggest you to use `sudo` because Docker Rootless runs `slirp4netns` with the `--disable-host-loopback` option, so the Gatling cannot access localhost.

## How to build the image

``` bash
docker build -t gatling-on-docker -f Dockerfile .
```

## How to run the container

1. Run the container using the following command:

``` bash
docker run -it -d --network=host --name gatling gatling-on-docker
```

Note: ensure to only remove `--network=host` if you know how to configure the `host.docker.internal`, or if you will call your machine's local ip, or if you are just not using Linux.

2. Execute the following command to start running the default simulation or rinha de backend simulation:

- The default simulation

``` bash
docker exec gatling /opt/gatling/bin/gatling.sh -rm local -sf /opt/gatling/user-files/simulations -s computerdatabase.ComputerDatabaseSimulation -rf /opt/gatling/results
```

- The rinha's simulation

``` bash
docker exec gatling /opt/gatling/bin/gatling.sh -rm local -sf /opt/gatling/user-files/simulations -s RinhaBackendCrebitosSimulation -rf /opt/gatling/results
```

## How to stop and remove the gatling container from your machine

``` bash
docker stop gatling && docker rm gatling
```


## How to run the container using a simulation from another Rinha (maybe Q2 or 2025?)

1. Start the container

``` bash
docker run -it -d -v /path/to/custom/config:/opt/gatling/conf \   
  -v /path/to/custom/user-files:/opt/gatling/user-files \   
  -v /path/to/results:/opt/gatling/results \   
  --network=host --name galing gatling-on-docker
```

2. Execute your simulation

``` bash
docker exec gatling /opt/gatling/bin/gatling.sh -rm local -sf /opt/gatling/user-files/simulations -s <name_of_your_simulation> -rf /opt/gatling/results
```
