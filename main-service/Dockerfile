FROM golang:1.23 AS builder

COPY . /app/
WORKDIR /app

RUN CGO_ENABLED=0 go build -ldflags "-X main.Environment=production" -o ./bin/app main.go

#--------------------------------------------------------------------------------------------
FROM scratch AS runtime

WORKDIR /app

#COPY --from=builder /app/static/ static/
COPY --from=builder /app/bin/app app
COPY --from=builder /app/optimizer/repo/csv/ optimizer/repo/csv/


EXPOSE 8080

ENTRYPOINT ["./app"]
