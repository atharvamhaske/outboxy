FROM golang:1.23 AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/orders ./orders && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/dispatcher ./dispatcher

FROM alpine:3.20

RUN adduser -D -g '' app
USER app
WORKDIR /home/app

COPY --from=builder /out/orders /usr/local/bin/orders
COPY --from=builder /out/dispatcher /usr/local/bin/dispatcher

ENV PATH="/usr/local/bin:${PATH}"

# Default to running the orders service; override in docker-compose for dispatcher
ENTRYPOINT ["orders"]


