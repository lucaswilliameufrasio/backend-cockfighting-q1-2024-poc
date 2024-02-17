FROM mcr.microsoft.com/openjdk/jdk:21-ubuntu

# working directory for gatling
WORKDIR /opt

# Gating version
ENV GATLING_VERSION 3.10.3

RUN apt-get update && apt-get install curl unzip git -y

# create directory for gatling install
RUN mkdir -p gatling

# install gatling
RUN mkdir -p /tmp/downloads && \
  curl -sf -o /tmp/downloads/gatling-$GATLING_VERSION.zip \
  # -L https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/3.10.3/gatling-charts-highcharts-bundle-3.10.3-bundle.zip && \
  -L https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/$GATLING_VERSION/gatling-charts-highcharts-bundle-$GATLING_VERSION-bundle.zip && \
  mkdir -p /tmp/archive && cd /tmp/archive && \
  unzip /tmp/downloads/gatling-$GATLING_VERSION.zip && \
  mv /tmp/archive/gatling-charts-highcharts-bundle-$GATLING_VERSION/* /opt/gatling/

# Add custom simulation
RUN git clone --single-branch --quiet https://github.com/zanfranceschi/rinha-de-backend-2024-q1.git && \
  ls -la rinha-de-backend-2024-q1/ && \
  ls -la rinha-de-backend-2024-q1/load-test/user-files/simulations/ && \
  mv rinha-de-backend-2024-q1/load-test/user-files/simulations/* /opt/gatling/user-files/simulations/

RUN ls -la /opt/gatling/user-files/

RUN ls -la /tmp/downloads
RUN ls -la /tmp/archive
RUN ls -la /opt/gatling

# change context to gatling directory
WORKDIR  /opt/gatling

# set directories below to be mountable from host
VOLUME ["/opt/gatling/conf","/opt/gatling/results","/opt/gatling/user-files"]

# set environment variables
ENV PATH /opt/gatling/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GATLING_HOME /opt/gatling
