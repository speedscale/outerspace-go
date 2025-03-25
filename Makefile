.PHONY: test test-with-proxymock coverage coverage-html clean proxymock-setup proxymock-run proxymock-stop

test:
	go test -v ./...

test-with-proxymock:
	export http_proxy=socks5h://localhost:4140 && \
	export https_proxy=socks5h://localhost:4140 && \
	export SSL_CERT_FILE=~/.speedscale/certs/tls.crt && \
	go test -v ./...
	$(MAKE) proxymock-stop

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f coverage.out coverage.html
	rm -rf .speedscale/proxymock

proxymock-setup:
	mkdir -p .speedscale
	sh -c "$$(curl -Lfs https://downloads.speedscale.com/proxymock/install-proxymock)"
	@if [ -z "$$PROXYMOCK_API_KEY" ]; then \
		echo "Error: PROXYMOCK_API_KEY is not set."; \
		exit 1; \
	fi
	export PATH="$$HOME/.speedscale:$$PATH" && \
	proxymock init --api-key "$$PROXYMOCK_API_KEY"
	@echo "Proxymock setup completed successfully."

proxymock-run:
	export PATH="$$HOME/.speedscale:$$PATH" && \
	nohup proxymock run --service "http=18080" --service "https=18443" --dir ./proxymock > proxymock.log 2>&1 & \
	sleep 5
	@if ! pgrep -f "proxymock run" > /dev/null; then \
		echo "Error: Proxymock is NOT running!"; \
		cat proxymock.log; \
		exit 1; \
	fi
	@echo "Proxymock started successfully."

proxymock-stop:
	pkill -f "proxymock run" || true
	@echo "Proxymock stopped." 