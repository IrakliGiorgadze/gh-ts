# ---------- build stage ----------
FROM golang:1.24 AS build
WORKDIR /src

# Enable reproducible builds and static binary
ENV CGO_ENABLED=0 GO111MODULE=on

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the API
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---------- runtime stage ----------
FROM gcr.io/distroless/static-debian12
WORKDIR /app

# App env defaults (overridable)
ENV API_PORT=8080 \
    APP_ENV=prod

COPY --from=build /out/api /app/api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]