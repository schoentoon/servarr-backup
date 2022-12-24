FROM docker.io/library/golang as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/app ./cmd/servarr-backup/...

FROM docker.io/alpine

COPY --from=builder /bin/app /usr/bin/servarr-backup

COPY --from=docker.io/restic/restic /usr/bin/restic /usr/bin/restic

CMD [ "/usr/bin/servarr-backup" ]
