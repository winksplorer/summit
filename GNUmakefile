SUMMIT_VERSION = 0.4

GO ?= go
TAR ?= tar
SED ?= sed
MINIFIER ?= minify -q
UPX ?= upx -9 --lzma

PREFIX ?= /usr/local

GOFLAGS = -buildvcs=false -trimpath
GO_LDFLAGS = -s -w -buildid= -X main.BuildDate=$(shell date -I) -X main.Version=$(SUMMIT_VERSION)

ifeq ($(SMALL),1)
	GOFLAGS += -gcflags=all=-l
endif

.PHONY: backend frontend all sea clean install

all: frontend backend

clean:
	rm -rf summit frontend-dist

# backend server build with go
backend:
	@echo "     GO ($(GO)) summit"
	@$(GO) mod tidy && $(GO) build -o summit $(GOFLAGS) -ldflags="$(GO_LDFLAGS)"
	@echo "     UPX ($(UPX)) summit"
	@$(UPX) summit

# bundle + minify frontend
frontend:
	@echo "  MKDIR (mkdir) frontend-dist/js"
	@mkdir -p frontend-dist/js
	@echo " MINIFY ($(MINIFIER)) frontend/js/*.js (exclude js/page, js/lib/page & js/independent) -> frontend-dist/js/bundle.min.js"
	@(printf 'frontend/js/main/core.js\0'; \
		find frontend/js -name '*.js' \
		! -path 'frontend/js/page/*' \
		! -path 'frontend/js/lib/page/*' \
		! -path 'frontend/js/independent/*' \
		! -path 'frontend/js/main/core.js' -print0 | sort -z) | xargs -0 cat | $(MINIFIER) --type=application/javascript > frontend-dist/js/bundle.min.js
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

# installs to prefix
install:
	install -m 755 summit $(PREFIX)/bin