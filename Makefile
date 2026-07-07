VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := fcc-drs
CMD     := ./cmd/fcc-drs

KATEX_VERSION := 0.16.11
HTMX_VERSION  := 2.0.4
MARKED_VERSION := 14
BULMA_VERSION := 1.0.4

VENDOR := static/vendor
DB     := data/requests.db

.PHONY: build build-dev run reseed deploy-staging deploy-prod assets clean

build:
	go build -tags prod -ldflags "-X main.version=$(VERSION)" -o $(BINARY) $(CMD)

build-dev:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) $(CMD)

run:
	DEV_MODE=TRUE go run $(CMD)

reseed:
	rm -f $(DB)
	@DEV_MODE=TRUE go run $(CMD) >/dev/null 2>&1 & \
	  APP_PID=$$!; \
	  until sqlite3 $(DB) "SELECT 1 FROM sqlite_master WHERE type='table' AND name='users'" 2>/dev/null | grep -q 1; \
	  do sleep 0.1; done; \
	  sqlite3 $(DB) < scripts/seed.sql; \
	  kill $$APP_PID 2>/dev/null; \
	  echo "✓ DB reseeded — run 'make run' to start."

deploy-staging:
	oc apply -f openshift/overlays/staging/secret.yaml
	oc apply -k openshift/overlays/staging

deploy-prod:
	oc apply -f openshift/overlays/prod/secret.yaml
	oc apply -k openshift/overlays/prod

assets:
	mkdir -p $(VENDOR)/katex/fonts $(VENDOR)/fonts
	@echo "→ Downloading HTMX $(HTMX_VERSION)..."
	curl -sL -o $(VENDOR)/htmx.min.js \
	  https://unpkg.com/htmx.org@$(HTMX_VERSION)/dist/htmx.min.js
	@echo "→ Downloading marked $(MARKED_VERSION)..."
	curl -sL -o $(VENDOR)/marked.min.js \
	  https://cdn.jsdelivr.net/npm/marked@$(MARKED_VERSION)/marked.min.js
	@echo "→ Downloading Bulma $(BULMA_VERSION)..."
	curl -sL -o $(VENDOR)/bulma.min.css \
	  https://cdn.jsdelivr.net/npm/bulma@$(BULMA_VERSION)/css/bulma.min.css
	@echo "→ Downloading KaTeX $(KATEX_VERSION)..."
	$(eval TMP := $(shell mktemp -d))
	curl -sL https://registry.npmjs.org/katex/-/katex-$(KATEX_VERSION).tgz \
	  | tar -xz -C $(TMP)
	cp $(TMP)/package/dist/katex.min.css              $(VENDOR)/katex/
	cp $(TMP)/package/dist/katex.min.js               $(VENDOR)/katex/
	cp $(TMP)/package/dist/contrib/auto-render.min.js $(VENDOR)/katex/
	cp $(TMP)/package/dist/fonts/*.woff2              $(VENDOR)/katex/fonts/
	rm -rf $(TMP)
	@echo "→ Downloading Inter font..."
	curl -sL -o $(VENDOR)/fonts/inter-cyrillic-ext.woff2  https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2JL7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-cyrillic.woff2      https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa0ZL7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-greek-ext.woff2     https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2ZL7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-greek.woff2         https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa1pL7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-vietnamese.woff2    https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa2pL7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-latin-ext.woff2     https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa25L7SUc.woff2
	curl -sL -o $(VENDOR)/fonts/inter-latin.woff2         https://fonts.gstatic.com/s/inter/v20/UcC73FwrK3iLTeHuS_nVMrMxCp50SjIa1ZL7.woff2
	@echo "✓ All assets ready."

clean:
	rm -f $(BINARY)
