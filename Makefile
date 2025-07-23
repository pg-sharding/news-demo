.PHONY: demo build clean test demo-api

demo: clean build
	@echo "🚀 Starting SPQR Demo..."
	docker compose up -d
	@echo "⏳ Waiting for services..."
	./scripts/wait-for-ready.sh
	@echo "🔧 Setting up database..."
	./scripts/setup-demo.sh
	@echo "✅ Demo ready! Try: make test"

build:
	go build -v ./cmd/apiserver
	go build -v ./cmd/rssparser

test:
	@echo "🔍 Testing data distribution..."
	./scripts/show-distribution.sh

demo-api:
	@echo "🌐 Testing API endpoints..."
	./scripts/test-api.sh

clean:
	docker compose down -v
	rm -f apiserver rssparser

.DEFAULT_GOAL := demo