GO ?= go
CC ?= clang

# Alpine is weird and wants us to directly link libarchive in SEA for some reason
ifneq ($(wildcard /usr/lib/libarchive.so),)
    LIBARCHIVE_FLAGS = /usr/lib/libarchive.so
else
    LIBARCHIVE_FLAGS =
endif

.PHONY: backend all sea

all: backend sea

clean:
	rm -f summit-server summit summit.tar.gz summit.tar.gz.o

backend:
	cd backend && go mod tidy && $(GO) build -o ../summit-server

sea:
	tar -czf summit.tar.gz summit-server frontend summit.crt summit.key
	ld -r -b binary -o summit.tar.gz.o summit.tar.gz
	clang -larchive -o summit sea.c summit.tar.gz.o $(LIBARCHIVE_FLAGS)

tlskey:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout summit.key -out summit.crt -subj "/C=US/ST=Washington/O=winksplorer & contributors/CN=summit ($(shell uname -n))"

install:
	install -m 755 summit /usr/bin