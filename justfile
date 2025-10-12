###################################
# Basic configuration
###################################

# For windows compatibility
set windows-shell := ["C:\\Program Files\\Git\\bin\\sh.exe", "-c"]

# Ignore recipe lines beginning with #.
set ignore-comments := true

# Load environment variables from .env file
set dotenv-load := true

# Format justfile
just-fmt:
    just --fmt --unstable

###################################
# Update
###################################

# Update Go
# To update golangci-lint, visit https://golangci-lint.run/docs/welcome/install/#binaries
# To update tools, Ctrl+Shift+P and search for "Go: Install/Update Tools"
update-go:
    go version
    winget upgrade GoLang.Go || true

# Update Go version in go.mod
update-mod:
    go mod edit -go=1.25.1

# Update dependencies
update:
    go mod tidy
    go get -t -u ./...
    go mod tidy

###################################
# Formatter and Linter
###################################

# Fmt
fmt:
    golangci-lint fmt

# Lint
lint:
    just fmt
    golangci-lint run --fix

###################################
# Run
###################################

# Run the project
run:
    ENV=dev go run ./src/.

###################################
# Tests
###################################

# Run tests
# -v: verbose
# -count=1: disable test cache
test:
    go test -count=1 ./test -v

###################################
# Dependencies
###################################

# Add dependency to go.mod
add package:
    go get {{ package }}

# Install binary package globally
# To delete an installed package, visit `C:\Users\xxx\go\bin`, and delete the exe file
# To check installed packages, visit the same folder
# To update an installed package, run `go install package@latest` again
# To update vscode go extension, Ctrl+Shift+P and search for "Go: Install/Update Tools"
install package:
    go install {{ package }}

###################################
# Build
###################################

# Build binary for Linux
build:
    GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o bin/main ./src

###################################
# Docker
###################################

# Build Docker image
docker:
    docker build -t latex .

# Stop and remove Docker container
stop:
    docker stop latex || true
    docker rm latex || true

# Run Docker container
container:
    just stop
    docker run -p 3001:3001 --name latex latex

# Run all steps
all:
    just lint
    just update
    just build
    just docker
    just container

###################################
# Artifact Registry
###################################

# Image path
IMAGE := env_var('REGION') + '-docker.pkg.dev/' + env_var('PROJECT_ID') + '/' + env_var('REPO') + '/' + env_var('PACKAGE')

# Setup tag
tag:
    docker tag latex {{ IMAGE }}

# Push image to Artifact Registry
push:
    docker push {{ IMAGE }}:latest

# List images in Artifact Registry
list:
    gcloud artifacts docker images list {{ IMAGE }}

# Delete image from Artifact Registry
delete:
    gcloud artifacts docker images delete {{ IMAGE }} --quiet

###################################
# Cloud Run
###################################

# Deploy to Cloud Run
deploy:
    gcloud run deploy $PACKAGE \
    --image {{ IMAGE }}:latest \
    --project $PROJECT_ID \
    --region $REGION \
    --allow-unauthenticated \
    --no-cpu-boost \
    --cpu=1 \
    --memory=256Mi \
    --timeout=30 \
    --concurrency=5 \
    --max-instances=5 \
    --port=8080 \
    --set-env-vars=GOMEMLIMIT=200MiB
