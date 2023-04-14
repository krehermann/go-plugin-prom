FROM prom/prometheus as prom
COPY tools/docker/prometheus_embedded.yml /etc/prometheus/prometheus.yml

# Build image
FROM golang:1.20-buster as buildgo
RUN go version
WORKDIR /usr/local/myapp


RUN apt-get -y update; apt-get -y install curl iproute2 lsof
COPY Makefile  ./
COPY tools/docker/run.sh ./
ADD go.mod go.sum ./
RUN go mod download

COPY common common
COPY cmd cmd
COPY api api

COPY server server
COPY plugin plugin

# Build the golang binary
RUN make all

# hacking... goal: prom running in the same container as my app
COPY --from=prom /prometheus /prometheus
COPY --from=prom /bin/prometheus /bin/prometheus
COPY --from=prom /bin/promtool /bin/promtool
COPY --from=prom /etc/prometheus/prometheus.yml /etc/prometheus/prometheus.yml


#RUN apt-get -y update; apt-get -y install curl iproute2 lsof
#WORKDIR /usr/local/myapp
#COPY --from=buildgo /workdir /usr/local/myapp/

#FROM ubuntu:20.04
#WORKDIR /usr/local/myapp
#RUN apt-get -y update; apt-get -y install curl

#COPY --from=buildgo /workdir /usr/local/myapp/



# prom metrics endpoints
EXPOSE 2112
# plugin range
EXPOSE 2113-2200
EXPOSE 9192
EXPOSE 9090



#ENTRYPOINT ["bin/cli"]

#HEALTHCHECK CMD curl -f http://localhost:6688/health || exit 1

CMD ["/bin/bash", "-c", "./run.sh"]