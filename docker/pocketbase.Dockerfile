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
FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR /skrump

COPY --from=build-stage /dist/ ./

VOLUME ./pb_data

EXPOSE 8090

USER nonroot:nonroot


ENTRYPOINT ["./pocketbase", "serve"]
