FROM golang:1.16-alpine as build

WORKDIR /app

COPY . .

WORKDIR /app/cmd

RUN --mount=type=cache,target=/root/.cache CGO_ENABLED=0 go build -v -ldflags '-w -s' -o main

# start with base image
FROM mysql:8.0.23 as build2

# import data into container
# All scripts in docker-entrypoint-initdb.d/ are automatically executed during container startup
COPY ./user/repository/mysql/*.sql /docker-entrypoint-initdb.d/
# Run stage
FROM alpine:latest 

WORKDIR /app

COPY --from=build /app/cmd .
CMD ["./main"]