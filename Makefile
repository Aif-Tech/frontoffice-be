.PHONY: test
test:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

.PHONY: cover
cover:
	go tool cover -html=coverage.out
