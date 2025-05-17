GO ?= go
CC ?= clang

SUMMIT_VERSION = 0.3

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
	cd backend && go mod tidy && $(GO) build -o ../summit-server -ldflags "-X main.BuildDate=$(shell date +%Y-%b-%d) -X main.Version=$(SUMMIT_VERSION)"
	strip summit-server

sea:
	tar -czf summit.tar.gz summit-server frontend
	ld -r -b binary -o summit.tar.gz.o summit.tar.gz
	clang -larchive -o summit sea.c summit.tar.gz.o $(LIBARCHIVE_FLAGS)

install:
	install -m 755 summit /usr/bin

bsdinstall:
	install -m 755 summit /usr/local/bin