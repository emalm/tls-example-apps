.PHONY: all backend frontend

all: backend frontend

dirs:
	mkdir -p bin/linux
	mkdir -p bin/darwin

backend: backend/main.go
	GOOS=linux go build -o bin/linux/backend ./backend
	GOOS=darwin go build -o bin/darwin/backend ./backend

frontend: frontend/main.go
	GOOS=linux go build -o bin/linux/frontend ./frontend
	GOOS=darwin go build -o bin/darwin/frontend ./frontend

