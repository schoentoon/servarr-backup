FROM docker.io/library/golang as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/app ./cmd/servarr-backup/...

FROM gcr.io/distroless/base

COPY --from=builder /bin/app /bin/app

CMD [ "/bin/app" ]
