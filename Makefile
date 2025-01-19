GO = go

.PHONY: backend all

all: backend

backend:
	cd backend && go mod tidy && $(GO) build -o ../summit