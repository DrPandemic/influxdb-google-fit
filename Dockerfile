ARG ARCH=
FROM ${ARCH}golang:alpine

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["influxdb_google_fit"]
