GOCMD=go

.PHONY: lint
lint:
	${GOCMD} run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

.PHONY: test
test:
	${GOCMD} test ./... -v -race -count=1

.PHONY: watch
watch:
	${GOCMD} run github.com/cespare/reflex --decoration=none --inverse-regex='^build/|^vendor/' make testsum

.PHONY: testsum
testsum:
	${GOCMD} run gotest.tools/gotestsum --format dots --raw-command -- go test --short --json ./...

.PHONY: build
build:
	${GOCMD} build -o gomo