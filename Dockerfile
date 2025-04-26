# BUILDER
FROM golang:1.24-alpine as builder
WORKDIR /go/modules/hunterio-test-api
COPY . .
# Forklift: Generate the forklift.go file.
RUN go install github.com/solher/forklift@latest && forklift -extensions static.sql,tmpl.sql,lazy.sql > forklift.go
RUN apk add --no-cache make && make install-api

# RUNNER
FROM alpine:latest
COPY --from=builder /go/bin/hunterio-test-api /usr/local/bin/hunterio-test-api
ENTRYPOINT ["/usr/local/bin/hunterio-test-api"]
EXPOSE 8080
