.PHONY: all backend frontend

all: dirs backend frontend

dirs:
	mkdir -p bin/linux/backend
	mkdir -p bin/linux/frontend
	mkdir -p bin/darwin/backend
	mkdir -p bin/darwin/frontend

backend: backend/main.go
	GOOS=linux go build -o bin/linux/backend/backend ./backend
	GOOS=darwin go build -o bin/darwin/backend/backend ./backend

frontend: frontend/main.go
	GOOS=linux go build -o bin/linux/frontend/frontend ./frontend
	GOOS=darwin go build -o bin/darwin/frontend/frontend ./frontend

