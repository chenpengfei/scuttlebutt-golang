# https://github.com/hashicorp/terraform/blob/master/Makefile

VERSION?=v0.0.12
TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TESTARGS?=-gcflags=-l

default: install

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

commit:
	git add .
	npm run commit

lint:
	golint ./...
	golangci-lint run ./...

tools:
	go get golang.org/x/tools/cmd/stringer
	go get golang.org/x/tools/cmd/cover
	go get github.com/golang/mock/mockgen
	go get github.com/golang/lint
	go install github.com/golang/lint

download:
	go mod tidy
	go mod download

publish:
	jfrog rt go-publish ${jfrog_repo_deployment} $(VERSION) --url=${jfrog_url} --user=${domain_username} --password=${domain_passward}


# docker generates docker image for develop
docker: generate
	sh -c "'$(CURDIR)/scripts/docker/build.sh' $(VERSION)"

# bin generates the releaseable binaries
bin: generate
	@TF_RELEASE=1 TF_DEV=true sh -c "'$(CURDIR)/scripts/build.sh'"

# test runs the unit tests
test: generate mock
	go list $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=2m -parallel=4

jmeter:
	rm -fr test/jmeter/report
	jmeter -n -t jmeter/test.jmx -l jmeter/log.jtl -e -o jmeter/report

# e2etest runs the end-to-end tests against a generated Terraform binary
# The TF_ACC here allows network access, but does not require any special
# credentials since the e2etests use local-only providers such as "null".
e2etest: generate
	TF_ACC=1 go test $(TESTARGS) -mod=vendor -v ./command/e2etest

# testrace runs the race checker
testrace: fmtcheck generate
	TF_ACC= go test -mod=vendor -race $(TEST) $(TESTARGS)

sonar: coverprofile
	@sh -c "'$(CURDIR)/scripts/sonar.sh'"

cover: coverprofile
	go tool cover -html=coverage.out
	rm coverage.out

coverprofile: mock
	@go tool cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	go test $(TEST) $(TESTARGS) -coverprofile=coverage.out

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# generate runs `go generate` to build the dynamically generated
# source files, except the protobuf stubs which are built instead with
# "make protobuf".
generate: tools
	GOFLAGS=-mod=vendor go generate ./...

swagger:
	swag init --g ./internal/app/server/server.go -o ./api/swagger

# mock runs `mockgen` to generate mock interfaces from source file
mock:
	@sh -c "'$(CURDIR)/scripts/mock.sh'"

clean:
	rm -fr bin
	rm jmeter.log

# disallow any parallelism (-j) for Make. This is necessary since some
# commands during the build process create temporary files that collide
# under parallel conditions.
.NOTPARALLEL:

.PHONY: build vendor bin cover default e2etest fmt fmtcheck generate test testacc testrace tools website

