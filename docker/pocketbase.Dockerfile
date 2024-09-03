FROM golang:1.23 AS build-stage

WORKDIR /skrump

COPY go.mod go.sum ./
COPY ../src ./src
COPY ../pb_migrations ./pb_migrations
COPY ../pb_public ./pb_public
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -C /skrump/src -o ../../dist/pocketbase
RUN mkdir /dist/pb_data
RUN mv ./pb_public /dist/pb_public

# # Run the tests in the container
# FROM build-stage AS run-test-stage
# RUN go test -v ./...

# Deploy the application binary into a lean image
# Image needs socat and useradd capabilities
FROM debian:stable-slim AS deb-extract

COPY --from=build-stage /dist/ /skrump
COPY --chmod=777 docker/includes/server.sh /usr/local/sbin/server.sh

# Install socat and setup user account
RUN apt-get update && apt-get install -y socat && \
    useradd -G daemon --shell "/bin/sh" -mk /dev/null "skrump" && \
    echo "skrump ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers;

VOLUME ["/skrump/pb_data", "/var/run/host_docker.sock"]
EXPOSE 8090

# Launches socat for docker host redirect and pocketbase.
# Pocketbase eats all arguments passed into the container exec
ENTRYPOINT ["/usr/local/sbin/server.sh"]
# Amazing blog post and github repo demonstrating a socat host docker.sock redirect
# https://www.knusbaum.org/posts/revisiting-docker-in-devenv
# https://github.com/klnusbaum/kdevenv
