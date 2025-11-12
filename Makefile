.PHONY: proto deps build run clean test submodule-update docker-push

submodule-update:
	@echo "Updating git submodules..."
	git submodule update --init --recursive
	git submodule update --remote --merge

proto:
	@echo "Generating Go code from protobuf..."
	@mkdir -p pkg/api
	protoc --go_out=pkg/api --go_opt=paths=source_relative \
		--go_opt=Mcommon/common.proto=github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/common \
		--go_opt=Manalyzer/analyzer.proto=github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/analyzer \
		--go-grpc_out=pkg/api --go-grpc_opt=paths=source_relative \
		--go-grpc_opt=Mcommon/common.proto=github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/common \
		--go-grpc_opt=Manalyzer/analyzer.proto=github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/analyzer \
		-I backend-common/proto \
		backend-common/proto/analyzer/analyzer.proto \
		backend-common/proto/common/common.proto

deps:
	go mod download
	go mod tidy

build: submodule-update proto
	go build -o bin/analyzer ./cmd/analyzer

run: build
	@if [ -f .env ]; then \
		echo "Loading .env file..."; \
		export $$(cat .env | grep -v '^#' | xargs) && ./bin/analyzer; \
	else \
		./bin/analyzer; \
	fi

clean:
	rm -rf bin/
	rm -rf pkg/api/

test:
	go test -v ./...

docker-push:
	@echo "Cleaning Docker system..."
	docker system prune -a -f --volumes
	@echo "Building Docker image..."
	docker build  -f build/Dockerfile -t analyzer .
	@echo "Tagging image..."
	docker tag analyzer cr.yandex/crpkimlhn85fg9vjfj7l/analyzer:latest
	@echo "Pushing to Yandex Container Registry..."
	docker push cr.yandex/crpkimlhn85fg9vjfj7l/analyzer:latest
	@echo "Done!"

