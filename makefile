# https://github.com/hashicorp/terraform/blob/master/Makefile

VERSION?=v0.0.12
TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TESTARGS?=-gcflags=-l

default: test

install:
	# Install and run Commitizen locally
	npm install --save-dev commitizen
	# initialize the conventional changelog adapter
	npx commitizen init cz-conventional-changelog --save-dev --save-exact
	# Install commitlint cli and conventional config
	npm install --save-dev @commitlint/{config-conventional,cli}
	echo "module.exports = {extends: ['@commitlint/config-conventional']};" > commitlint.config.js
	# Install husky as devDependency, a handy git hook helper available on npm
	# This allows us to add git hooks directly into our package.json via the husky.hooks field
	npm install --save-dev husky

tools:
#	go get golang.org/x/tools/cmd/stringer
#	go get github.com/golang/mock/mockgen
#	go get golang.org/x/tools/cmd/cover
#   go get github.com/mattn/goveralls

download:
	go mod tidy
	go mod download

run:
	go run examples/scuttlebutt/main.go

benchmark:
	go test -bench=. examples/pull-stream/random_test.go examples/pull-stream/random.go

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run ./...

test: generate mock
	go list $(TEST) | xargs -t -n4 go test -race $(TESTARGS) -timeout=2m -parallel=4

coverprofile: mock
	@go tool cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	go test -race $(TEST) $(TESTARGS) -coverprofile=coverage.out -covermode=atomic

cover: coverprofile
	go tool cover -html=coverage.out
	rm coverage.out

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# generate runs `go generate` to build the dynamically generated
# source files, except the protobuf stubs, which are built instead with
# "make protobuf".
generate: tools
	GOFLAGS=-mod=vendor go generate ./...

# mock runs `mockgen` to generate mock interfaces from source file
mock:
	@sh -c "'$(CURDIR)/scripts/mock.sh'"

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: default

