## Setup Project
- Docker (recommended)
- Local

## Docker
- `docker-compose up --build -d`
- go to [localhost:3000](localhost:3000)

## Local
- git clone https://github.com/feverxai/go-rabbit-test.git
- setup following environment variables to your mysql
```
      DB_HOST: db
      DB_PORT: 3306
      DB_USERNAME: rabbit
      DB_PASSWORD: password
      DB_DATABASE: rabbit
```
- `cd app`
- `go mod download`
- `go run ./cmd/shorten-url`
- go to [localhost:3000](localhost:3000)

## Usage
Example of usage is in `shorten-url.postman_collection.json`