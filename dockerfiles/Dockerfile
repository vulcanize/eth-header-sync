FROM golang:1.13-alpine as builder

RUN apk --update --no-cache add make git g++

# Build statically linked binary (wonky path because of Dep)
WORKDIR /go/src/github.com/vulcanize/eth-header-sync
ADD . .
RUN GCO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' .

# Copy migration tool
WORKDIR /
ARG GOOSE_VER="v2.6.0"
ADD https://github.com/pressly/goose/releases/download/${GOOSE_VER}/goose-linux64 ./goose
RUN chmod +x ./goose

# Second stage
FROM alpine

WORKDIR /app

COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/eth-header-sync eth-header-sync
COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/environments/example.toml config.toml
COPY --from=builder /go/src/github.com/vulcanize/eth-header-sync/db/migrations migrations/vulcanizedb
COPY --from=builder /goose goose

CMD ["./startup_script.sh"]
