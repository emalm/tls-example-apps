all: backend frontend

dirs:
	mkdir -p backend/bin/linux
	mkdir -p frontend/bin/linux

backend: backend/main.go
	GOOS=linux go build -o backend/bin/linux/backend.linux ./backend

frontend: frontend/main.go
	GOOS=linux go build -o frontend/bin/linux/frontend.linux ./frontend

