FROM golang:1.18-alpine AS builder

WORKDIR /app
# ENV GO111MODULE=auto

COPY . ./

RUN go mod download && \
    go build -o /main


FROM golang:1.18-alpine

WORKDIR /

COPY --from=builder /app/ /main /

#EXPOSE 9000

CMD ["/main"]

