VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY  := fcc-drs
CMD     := ./cmd/fcc-drs

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
	@echo "→ Installing frontend libraries via npm..."
	npm ci
	./scripts/vendor-assets.sh $(VENDOR)
	@echo "✓ All assets ready."

clean:
	rm -f $(BINARY)
