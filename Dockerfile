FROM golang:1.16-buster

EXPOSE 1001

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN go build travis .

CMD ["/app/travis"]