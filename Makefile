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
	rm -f summit-server summit summit.tar.xz summit.tar.xz.o

backend:
	cd backend && go mod tidy && $(GO) build -o ../summit-server -ldflags "-X main.BuildDate=$(shell date +%Y-%b-%d) -X main.Version=$(SUMMIT_VERSION)"
	strip summit-server

sea:
	tar -cJf summit.tar.xz summit-server frontend
	ld -r -b binary -o summit.tar.xz.o summit.tar.xz
	clang -larchive -o summit sea.c summit.tar.xz.o $(LIBARCHIVE_FLAGS)

install:
	install -m 755 summit /usr/bin

bsdinstall:
	install -m 755 summit /usr/local/bin