FROM golang:alpine

MAINTAINER Dean Sheather <dean@deansheather.com>

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o cloudbleed cloudbleed.go

EXPOSE 8080
CMD ["/app/cloudbleed"]
