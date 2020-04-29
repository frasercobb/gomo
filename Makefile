GOCMD=go

.PHONY: ci
ci: unit lint

.PHONY: lint
lint:
	${GOCMD} run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: lint-fix
lint-fix:
	${GOCMD} run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

.PHONY: watch-lint
watch-lint:
	${GOCMD} run github.com/cespare/reflex --decoration=none --inverse-regex='^build/|^vendor/' make lint

.PHONY: unit
unit:
	${GOCMD} test ./... -v -race -count=1 --short

.PHONY: watch
watch:
	${GOCMD} run github.com/cespare/reflex --decoration=none --inverse-regex='^build/|^vendor/' make testsum

.PHONY: testsum
testsum:
	${GOCMD} run gotest.tools/gotestsum --format dots --raw-command -- go test -v --short --json ./...

.PHONY: build
build:
	${GOCMD} build -o ./build-output/gomo

.PHONY: run
run:
	${GOCMD} run .