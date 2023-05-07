FROM golang

RUN go version
ENV GOPATH=/

COPY ./ ./

# Обновление пакетов
RUN apt-get update

# Установка postgresql-client
RUN apt-get -y install postgresql-client

# Запуск wait-for-postgres.sh
RUN chmod +x wait-for-postgres.sh

# Установка всех зависимостей
RUN go mod download

# Сборка приложения
RUN go build -o server-app-main ./cmd/main.go

# Запуск приложения
CMD ["./server-app-main"]