################################################################################
# BUILDER IMAGE
################################################################################

FROM golang:1.24 AS builder
ARG VERSION
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# ------------------------------------------------------------------------------
# Define and create working directory
# ------------------------------------------------------------------------------

ENV APP_HOME=/app
RUN mkdir -p ${APP_HOME}
WORKDIR ${APP_HOME}

# ------------------------------------------------------------------------------
# Copy source code
# ------------------------------------------------------------------------------

COPY . .

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------

# Set target architecture based on TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
      "linux/amd64") export GOARCH=amd64 ;; \
      "linux/arm64") export GOARCH=arm64 ;; \
      *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build -a -o /outerspace-go \
    -ldflags "-w -X main.Version=${VERSION} -X main.BuildTime=$(TZ=UTC date +%Y-%m-%dT%H:%M:%S%z)"

################################################################################
# RELEASE IMAGE
################################################################################

FROM debian:bookworm-slim

# ------------------------------------------------------------------------------
# Install persistent dependencies
# ------------------------------------------------------------------------------

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# ------------------------------------------------------------------------------
# Define working directory & user
# ------------------------------------------------------------------------------

ENV APP_HOME=/app
ARG APP_USER=1001
WORKDIR ${APP_HOME}

# ------------------------------------------------------------------------------
# Copy built application
# ------------------------------------------------------------------------------

COPY --from=builder --chown=${APP_USER} /outerspace-go ./outerspace-go
RUN chmod +x ./outerspace-go

# ------------------------------------------------------------------------------
# Define user, exposed ports, additional env vars, entrypoint and cmd
# ------------------------------------------------------------------------------

USER ${APP_USER}
EXPOSE 8443 8080
CMD ["/app/outerspace-go"]
