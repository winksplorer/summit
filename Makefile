GO ?= go
CC ?= clang

.PHONY: backend all sea

all: backend sea

clean:
	rm server summit summit.tar.gz summit.tar.gz.o

backend:
	cd backend && go mod tidy && $(GO) build -o ../server

sea:
	tar -czf summit.tar.gz server frontend
	ld -r -b binary -o summit.tar.gz.o summit.tar.gz
	clang -larchive -o summit sea.c summit.tar.gz.o