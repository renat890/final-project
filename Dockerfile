FROM golang:1.24 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-app

FROM alpine
ARG TODO_PORT
WORKDIR /
COPY --from=build /todo-app /todo-app
COPY web ./web
EXPOSE 8080
ENTRYPOINT [ "/todo-app" ]