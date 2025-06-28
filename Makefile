
.PHONY: build test

APP_NAME = AppNameHere
VERSION = $(shell rg -No "\d+\.\d+\.\d+" internal/about/version.go)
MAKE_DEPS = gh go git rg docker codecov
USER = arustydev
GHCR = github.com/$(USER)/$(APP_NAME)
DOCKER_CR = docker.io/$(USER)/$(APP_NAME)

build:
	@go build -o cmd/cli/$(APP_NAME) cmd/cli/main.go
	@go build -o cmd/tui/$(APP_NAME) cmd/tui/main.go
	@go build -o cmd/api/$(APP_NAME) cmd/api/main.go
	@go build -o cmd/headless/$(APP_NAME) cmd/headless/main.go

release: tag
	#Creating Release for GitHub
	@gh release create v$(VERSION) --title "v$(VERSION)" --notes-from-tag --fail-on-no-commits 'cmd/cli/$(APP_NAME)#CLI Binary' 'cmd/tui/$(APP_NAME)#TUI Binary' 'cmd/api/$(APP_NAME)#API Binary' 'cmd/headless/$(APP_NAME)#Headless Binary'

pre-publish: tag
	#Creating Release for GitHub
	@gh release create v$(VERSION) --title "v$(VERSION)" --draft --prerelease --notes-from-tag 'cmd/cli/$(APP_NAME)#CLI Binary' 'cmd/tui/$(APP_NAME)#TUI Binary' 'cmd/api/$(APP_NAME)#API Binary' 'cmd/headless/$(APP_NAME)#Headless Binary'

build-release:
	#Building CLI
	@docker build --tag $(APP_NAME)-cli:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg NAME=$(APP_NAME) --target cli ./build/
	#Building TUI
	@docker build --tag $(APP_NAME)-tui:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg NAME=$(APP_NAME) --target tui ./build/
	#Building API
	@docker build --tag $(APP_NAME)-api:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg NAME=$(APP_NAME) --target api ./build/
	#Building Headless
	@docker build --tag $(APP_NAME)-headless:$(VERSION) --build-arg VERSION=$(VERSION) --build-arg NAME=$(APP_NAME) --target headless ./build/

tag: build-release
	#Tagging
	@git tag -a $(VERSION)
	@git push origin tag $(VERSION)
	@docker image tag $(APP_NAME):$(VERSION) $(GHCR):$(VERSION)
	@docker image push $(GHCR):$(VERSION)
	@docker image tag $(APP_NAME):$(VERSION) $(DOCKER_CR):$(VERSION)
	@docker image push $(DOCKER_CR):$(VERSION)

test:
	# Note: The exact path to policy.go might vary. A more robust way is to define an interface in your code
	# that mirrors azcore.TokenCredential if you have trouble with direct generation.
	# Or, more commonly, your code will accept an azcore.TokenCredential interface.
	@mockgen -source=$(GOPATH)/pkg/mod/github.com/Azure/azure-sdk-for-go/sdk/azcore@vX.Y.Z/policy/policy.go -destination=mocks/mock_azcore_policy.go -package=mocks TokenCredential
	@go test -v ./...

upgrade-make-deps:
	@echo "TODO: Script upgrading make deps"
	@#brew update
	@#brew upgrade

install-make-deps:
	@echo "TODO: Script installing make deps"
	@#pip install codecov-cli --break-system-packages
	@#brew install ripgrep gh git go orbstack golangci-lint goimports

clean:
	@rm -rf .credentials
	@rm -rf ./test/.credentials
	@rm -f cmd/cli/$(APP_NAME)
	@rm -f cmd/tui/$(APP_NAME)
	@rm -f cmd/api/$(APP_NAME)
	@rm -f cmd/headless/$(APP_NAME)
	@docker image rm $(shell docker image inspect $(APP_NAME)-cli:$(VERSION) | jq '.[].Id' | rg -o "sha256:(.{12})" | cut -d: -f2)
	@docker image rm $(shell docker image inspect $(APP_NAME)-tui:$(VERSION) | jq '.[].Id' | rg -o "sha256:(.{12})" | cut -d: -f2)
	@docker image rm $(shell docker image inspect $(APP_NAME)-api:$(VERSION) | jq '.[].Id' | rg -o "sha256:(.{12})" | cut -d: -f2)
	@docker image rm $(shell docker image inspect $(APP_NAME)-headless:$(VERSION) | jq '.[].Id' | rg -o "sha256:(.{12})" | cut -d: -f2)
