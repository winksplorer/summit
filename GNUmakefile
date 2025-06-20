SUMMIT_VERSION = 0.3

GO ?= go
TAR ?= tar
SED ?= sed
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

# safeguard against log.Panicln midnight moments
check_panic:
	@echo "  CHECK (grep) .panic"
	@! grep -rIi '\.panic' backend/ || (echo 'learn to read your code!!' && exit 1)

# backend server build with go (and also run panic safeguard)
backend: check_panic
	@echo "     GO (${GO}) backend -> summit-server"
	@cd backend && $(GO) mod tidy && $(GO) build -o ../summit-server -ldflags="-s -w -X main.BuildDate=$(shell date +%Y-%b-%d) -X main.Version=$(SUMMIT_VERSION)"

# bundle + minify frontend
frontend:
	@echo "  MKDIR (mkdir) frontend-dist/js"
	@mkdir -p frontend-dist/js
	@echo " MINIFY ($(MINIFIER)) frontend/js/*.js (exclude js/page, js/lib/page & js/independent) -> frontend-dist/js/bundle.min.js"
	@(printf 'frontend/js/main/core.js\0'; find frontend/js -name '*.js' \
		! -path 'frontend/js/page/*' \
		! -path 'frontend/js/lib/page/*' \
		! -path 'frontend/js/independent/*' \
		! -path 'frontend/js/main/core.js' -print0) | xargs -0 cat | $(MINIFIER) --type=application/javascript > frontend-dist/js/bundle.min.js
	@echo "   COPY ($(TAR)) frontend (exclude js/main & js/lib, BUT include js/lib/page) -> frontend-dist"
	@cd frontend && find . -type f \
		\( -path './js/main*' -o -path './js/lib*' \) \
		! -path './js/lib/page*' -prune -o -type f -print | \
		$(TAR) --no-recursion -cT - | $(TAR) -x -C ../frontend-dist
	@echo " MINIFY ($(MINIFIER)) frontend-dist/css"
	@find frontend-dist/css -type f -name "*.css" -exec $(MINIFIER) --type=css {} -o {} \;
	@echo " MINIFY ($(MINIFIER)) frontend-dist/js"
	@find frontend-dist/js -type f ! -name "*.min.js" -exec $(MINIFIER) --type=js {} -o {} \;
	@echo "REPLACE (${SED}) js_bundle markers"
	@$(SED) -i '/<!-- JS_BUNDLE_START -->/,/<!-- JS_BUNDLE_END -->/c\<script src="js/bundle.min.js" defer></script>' frontend-dist/template/base.html
	@echo "REPLACE (${SED}) remove markers"
	@find frontend-dist/template -type f ! -name "base.html" -exec $(SED) -i '/<!-- REMOVE_MARKER_START -->/,/<!-- REMOVE_MARKER_END -->/d' {} +

# compress frontend+backend, then embeds it into a binary combined with compiled SEA
sea:
	@echo "     XZ ($(TAR)) summit-server + frontend-dist -> summit.tar.xz"
	@$(TAR) -cJf summit.tar.xz summit-server frontend-dist
	@echo "     LD ($(LD)) summit.tar.xz -> summit.tar.xz.o"
	@$(LD) -r -b binary -o summit.tar.xz.o summit.tar.xz
	@echo "     CC ($(CC)) sea.c + summit.tar.xz.o -> summit"
	@$(CC) -o summit sea.c summit.tar.xz.o $(LIBARCHIVE_FLAGS) -larchive

install:
	install -m 755 summit $(PREFIX)/bin