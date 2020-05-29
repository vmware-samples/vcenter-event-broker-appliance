COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell /bin/cat ../VERSION)
IMAGE_REPO=vmware
IMAGE_NAME=$(IMAGE_REPO)/veba-event-router
DIST_FOLDER=dist
BINARY=vmware-event-router
BUILD_TAG=$(IMAGE_NAME):$(COMMIT)
VERSION_TAG=$(IMAGE_NAME):$(VERSION)
LATEST_TAG=$(IMAGE_NAME):latest

GIT_BRANCH := $(shell git branch --show-current)
GIT_NOT_CLEAN_CHECK := $(shell git status --porcelain)
export GO111MODULE=on

PKGS     = $(or $(PKG),$(shell env GO111MODULE=on go list ./...))
TESTPKGS = $(shell env GO111MODULE=on go list -f \
            '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' \
            $(PKGS))
TIMEOUT  = 20

.PHONY: release tidy build tag push output vendor test

default: build

release: build tag push output
	$(info Make: Making new release for image "$(IMAGE_NAME)" based on commit "$(COMMIT).)

vendor:
	$(info Make: syncing Go dependencies with vendor folder.)
	go mod vendor -v

tidy:
	$(info Make: syncing and cleaning up Go dependencies.)
	go mod tidy -v

# intended for local dev use (won't check for unclean git)
binary: test tidy
	$(info Make: Building binary "$(DIST_FOLDER)/$(BINARY)".)
	CGO_ENABLED=0 go build -a -installsuffix nocgo -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT}" -o $(DIST_FOLDER)/$(BINARY) cmd/main.go

build: test tidy
	$(info Make: Building image "$(IMAGE_NAME)".)
	$(if $(GIT_NOT_CLEAN_CHECK), $(error "Dirty Git repository!"))
	docker build -t $(LATEST_TAG) --build-arg COMMIT=$(COMMIT) --build-arg VERSION=$(VERSION) .

gofmt:
	$(info Make: Checking code is gofmted.)
	@test -z "$(shell gofmt -s -l -d -e ./cmd | tee /dev/stderr)"

test: gofmt
	GORACE=history_size=5 go test -race -timeout $(TIMEOUT)s -cover $(TESTPKGS) 

tag: 
	$(info Make: Tagging image "$(LATEST_TAG)" with "$(BUILD_TAG)" and "$(VERSION_TAG)".)
	@docker tag $(LATEST_TAG) $(BUILD_TAG) 
	@docker tag $(LATEST_TAG) $(VERSION_TAG)

push: tag
	$(info Make: Pushing image "$(IMAGE_NAME)".)
ifneq ($(GIT_BRANCH),master)
	$(error "Not on master branch, won't push!")
endif

	docker push $(BUILD_TAG)
	docker push $(VERSION_TAG)
	docker push $(LATEST_TAG)

output: test
	@echo Docker Image: $(BUILD_TAG)
	@echo Docker Image: $(VERSION_TAG)
	@echo Docker Image: $(LATEST_TAG)
