.PHONY: test coverage coverage-html clean proxymock-mock run build build-client integration-test load-test http-test http-test-recording bump-major bump-minor bump-patch version docker-build docker-build-client docker-buildx-setup docker-push-server docker-push-client docker-push tag release-patch release-minor release-major update-k8s

# Define proxymock environment variables
PROXYMOCK_ENV = http_proxy=socks5h://localhost:4140 \
                https_proxy=socks5h://localhost:4140 \
                SSL_CERT_FILE=~/.speedscale/certs/tls.crt

# Find most recent recording or snapshot directory
PROXYMOCK_RECORDING := $(shell find ./proxymock \( -name "recorded-*" -o -name "snapshot-*" -o -name "pulled-*" \) -type d -exec ls -td {} + | head -n 1)

# Version management
CURRENT_VERSION := $(shell cat VERSION)
VERSION_PARTS := $(subst ., ,$(subst v,,$(CURRENT_VERSION)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

build:
	go build -o outerspace-go -ldflags "-X main.Version=$(CURRENT_VERSION) -X main.BuildTime=$(shell TZ=UTC date +%Y-%m-%dT%H:%M:%S%z)" main.go

build-client:
	go build -o outerspace-client -ldflags "-X main.Version=$(CURRENT_VERSION) -X main.BuildTime=$(shell TZ=UTC date +%Y-%m-%dT%H:%M:%S%z)" ./cmd/client

run:
	go run main.go

clean:
	rm -f coverage.out coverage.html outerspace-go outerspace-client
	rm -rf logs
	rm -rf proxymock/mocked-* proxymock/replayed-*

integration-test: build proxymock-mock
	mkdir -p logs
	echo "Starting outerspace-go in background with proxymock..."
	$(PROXYMOCK_ENV) ./outerspace-go > logs/outerspace.log 2>&1 & echo $$! > logs/outerspace.pid
	echo "Waiting for outerspace-go to start..."
	sleep 2
	echo "Running integration tests with proxymock..."
	proxymock replay --in $(PROXYMOCK_RECORDING) --fail-if requests.response-pct!=100
	echo "Cleaning up..."
	-pkill -f "outerspace-go" || true
	-pkill -f "proxymock" || true
	echo "Integration tests completed. See logs in the logs directory."

load-test: build proxymock-mock
	mkdir -p logs
	echo "Starting outerspace-go in background with proxymock..."
	$(PROXYMOCK_ENV) ./outerspace-go > logs/outerspace.log 2>&1 & echo $$! > logs/outerspace.pid
	echo "Waiting for outerspace-go to start..."
	sleep 2
	echo "Running load tests with proxymock..."
	proxymock replay --in $(PROXYMOCK_RECORDING) --no-out --vus 10 --for 1m --fail-if "latency.p95 > 200"
	echo "Cleaning up..."
	-pkill -f "outerspace-go" || true
	-pkill -f "proxymock" || true
	echo "Load tests completed. See logs in the logs directory."

http-test: build
	-pkill -f outerspace-go || true
	mkdir -p logs
	echo "Starting outerspace-go in background..."
	./outerspace-go > logs/outerspace.log 2>&1 & 
	echo $! > logs/outerspace.pid
	echo "Waiting for outerspace-go to start..."
	sleep 2
	echo "Running http tests from tests/test.http..."
	./tests/run_http_tests.sh
	@echo "Cleaning up..."
	-pkill -f outerspace-go || true

http-test-recording: build
	-pkill -f outerspace-go || true
	-pkill -f "proxymock record" || true
	mkdir -p logs
	echo "Starting proxymock record in background..."
	nohup proxymock record > logs/proxymock-record.log 2>&1 & echo $$! > logs/proxymock-record.pid
	echo "Waiting for proxymock record to start..."
	sleep 2
	echo "Starting outerspace-go in background..."
	$(PROXYMOCK_ENV) ./outerspace-go > logs/outerspace.log 2>&1 & 
	echo $! > logs/outerspace.pid
	echo "Waiting for outerspace-go to start..."
	sleep 2
	echo "Running http tests from tests/test.http in recording mode..."
	./tests/run_http_tests.sh --recording
	@echo "Cleaning up..."
	-pkill -f outerspace-go || true
	-pkill -f "proxymock record" || true

proxymock-mock:
	mkdir -p logs
	nohup proxymock mock --in $(PROXYMOCK_RECORDING) --no-out > logs/proxymock-mock.log 2>&1 & \
	sleep 2
	@if ! pgrep -f "proxymock mock" > /dev/null; then \
		echo "Error: Proxymock is NOT mocking!"; \
		cat logs/proxymock-mock.log; \
		exit 1; \
	fi
	@echo "Proxymock started successfully."

version:
	@echo "Current version: $(CURRENT_VERSION)"

bump-patch:
	@echo "v$(MAJOR).$(MINOR).$$(expr $(PATCH) + 1)" > VERSION
	@echo "Version bumped to: $$(cat VERSION)"
	@$(MAKE) update-k8s

bump-minor:
	@echo "v$(MAJOR).$$(expr $(MINOR) + 1).0" > VERSION
	@echo "Version bumped to: $$(cat VERSION)"
	@$(MAKE) update-k8s

bump-major:
	@echo "v$$(expr $(MAJOR) + 1).0.0" > VERSION
	@echo "Version bumped to: $$(cat VERSION)"
	@$(MAKE) update-k8s

update-k8s:
	@sed -i '' 's/newTag: v[0-9]*\.[0-9]*\.[0-9]*/newTag: $(CURRENT_VERSION)/g' k8s/kustomization.yaml
	@sed -i '' 's/image: ghcr\.io\/speedscale\/outerspace-go-client:v[0-9]*\.[0-9]*\.[0-9]*/image: ghcr.io\/speedscale\/outerspace-go-client:$(CURRENT_VERSION)/g' k8s/client-deployment.yaml
	@sed -i '' 's/image: ghcr\.io\/speedscale\/outerspace-go:v[0-9]*\.[0-9]*\.[0-9]*/image: ghcr.io\/speedscale\/outerspace-go:$(CURRENT_VERSION)/g' k8s/server-deployment.yaml
	@echo "Updated k8s manifests to $(CURRENT_VERSION)"

docker-build:
	docker build --build-arg VERSION=$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go:$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go:latest .

docker-build-client:
	docker build --build-arg VERSION=$(CURRENT_VERSION) -f Dockerfile.client -t ghcr.io/speedscale/outerspace-go-client:$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go-client:latest .

docker-buildx-setup:
	docker buildx create --name outerspace-builder --use || docker buildx use outerspace-builder

docker-push-server: docker-buildx-setup
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go:$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go:latest --push .

docker-push-client: docker-buildx-setup
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(CURRENT_VERSION) -f Dockerfile.client -t ghcr.io/speedscale/outerspace-go-client:$(CURRENT_VERSION) -t ghcr.io/speedscale/outerspace-go-client:latest --push .

docker-push: docker-push-server docker-push-client

tag:
	git tag $(CURRENT_VERSION)
	git push origin $(CURRENT_VERSION)
	@echo "Tagged and pushed $(CURRENT_VERSION) to trigger CI build"

release-patch: bump-patch tag
	@echo "Released patch version: $$(cat VERSION)"

release-minor: bump-minor tag
	@echo "Released minor version: $$(cat VERSION)"

release-major: bump-major tag
	@echo "Released major version: $$(cat VERSION)"