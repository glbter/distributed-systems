FROM golang:1.23 AS builder

# RUN apt-get update \
#     && apt-get install -y --no-install-recommends apt-transport-https \
#     && go install github.com/a-h/templ/cmd/templ@latest \
#     && apt-get install make

# RUN curl -fsSL https://deb.nodesource.com/setup_22.x -o nodesource_setup.sh  \
#     && bash nodesource_setup.sh  \
#     && apt-get install -y --no-install-recommends nodejs

COPY . /app/
WORKDIR /app

# RUN npm i
RUN CGO_ENABLED=0 go build -ldflags "-X main.Environment=production" -mod=vendor -o ./bin/ourApp ./cmd/main.go

#--------------------------------------------------------------------------------------------
FROM scratch AS runtime

WORKDIR /app

COPY --from=builder /app/static/ static/
COPY --from=builder /app/bin/ourApp ourApp

EXPOSE 8080

ENTRYPOINT ["./ourApp"]
