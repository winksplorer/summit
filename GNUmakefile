SUMMIT_VERSION = 0.3

GO ?= go
CC ?= $(shell command -v clang >/dev/null 2>&1 && echo clang || echo gcc)
LD ?= ld
TAR ?= tar
MINIFIER ?= minify

PREFIX ?= /usr

# Alpine is weird and wants us to directly link libarchive in SEA for some reason
ifneq ($(wildcard /usr/lib/libarchive.so),)
    LIBARCHIVE_FLAGS = /usr/lib/libarchive.so
else
    LIBARCHIVE_FLAGS =
endif

.PHONY: backend frontend all sea clean install

all: backend frontend sea

clean:
	rm -rf summit-server summit summit.tar.xz summit.tar.xz.o frontend-dist

check_panic:
	@echo "  CHECK .panic"
	@! grep -rIi '\.panic' backend/ || (echo 'learn to read your code!!' && exit 1)


backend: check_panic
	@echo "     GO backend -> summit-server"
	@cd backend && $(GO) mod tidy && $(GO) build -o ../summit-server -ldflags="-s -w -X main.BuildDate=$(shell date +%Y-%b-%d) -X main.Version=$(SUMMIT_VERSION)"

frontend:
	@echo "  MKDIR frontend-dist/js"
	@mkdir -p frontend-dist/js
	@echo " MINIFY frontend/js/*.js (exclude js/page, js/lib/page & js/independent) -> frontend-dist/js/bundle.min.js"
	@(printf 'frontend/js/main/core.js\0'; find frontend/js -name '*.js' \
		! -path 'frontend/js/page/*' \
		! -path 'frontend/js/lib/page/*' \
		! -path 'frontend/js/independent/*' \
		! -path 'frontend/js/main/core.js' -print0) | xargs -0 cat | $(MINIFIER) --type=application/javascript > frontend-dist/js/bundle.min.js
	@echo "   COPY frontend (exclude js/main & js/lib, BUT include js/lib/page) -> frontend-dist"
	@cd frontend && find . -type f \
		\( -path './js/main*' -o -path './js/lib*' \) \
		! -path './js/lib/page*' -prune -o -type f -exec cp --parents {} ../frontend-dist/ \;
	@echo " MINIFY frontend-dist/css"
	@find frontend-dist/css -type f -name "*.css" -exec $(MINIFIER) --type=css {} -o {} \;
	@echo " MINIFY frontend-dist/js"
	@find frontend-dist/js -type f ! -name "*.min.js" -exec $(MINIFIER) --type=js {} -o {} \;
	@echo "REPLACE js_bundle markers"
	@sed -i '/<!-- JS_BUNDLE_START -->/,/<!-- JS_BUNDLE_END -->/c\<script src="js/bundle.min.js" defer></script>' frontend-dist/template/base.html
	@echo "REPLACE remove markers"
	@find frontend-dist/template -type f ! -name "base.html" -exec sed -i '/<!-- REMOVE_MARKER_START -->/,/<!-- REMOVE_MARKER_END -->/d' {} +

sea:
	@echo "     XZ summit-server + frontend-dist -> summit.tar.xz"
	@$(TAR) -cJf summit.tar.xz summit-server frontend-dist
	@echo "     LD summit.tar.xz -> summit.tar.xz.o"
	@$(LD) -r -b binary -o summit.tar.xz.o summit.tar.xz
	@echo "     CC sea.c + summit.tar.xz.o -> summit"
	@$(CC) -o summit sea.c summit.tar.xz.o $(LIBARCHIVE_FLAGS) -larchive

install:
	install -m 755 summit $(PREFIX)/bin