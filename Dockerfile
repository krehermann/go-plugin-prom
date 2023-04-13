# Build image
FROM golang:1.20-buster as buildgo
RUN go version
WORKDIR /working

COPY Makefile  ./

ADD go.mod go.sum ./
RUN go mod download

COPY common common
COPY cmd cmd
COPY api api

COPY server server
COPY plugin plugin

# Build the golang binary
RUN make all

# prom metrics endpoints
EXPOSE 2112
# plugin range
EXPOSE 2113-2200

ENTRYPOINT ["bin/cli"]

#HEALTHCHECK CMD curl -f http://localhost:6688/health || exit 1

CMD ["server", "start"]