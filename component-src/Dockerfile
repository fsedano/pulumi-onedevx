FROM golang:1.22 as BUILDER
WORKDIR /app
COPY go.* /app/
RUN go mod download
COPY *.go /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/app .

FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/app /app/
CMD ["/app/app"]