SHELL := $(shell which bash)
NAME := calc-hash
PACKAGE_NAME := $(shell cat go.mod |grep "^module" |cut -d' ' -f2)

OUR_PACKAGES := $(shell go list ./... | grep -v '/vendor/')
GOX := gox
BUILT:=$(shell date -u +%Y-%m-%dT%H:%M:%S%z)

GIT_COMMIT := $(shell git rev-parse --short HEAD)
LAST_TAG ?= $(shell git log --decorate --no-color --pretty="format:%d" |awk 'match($$0, "[(]?tag:\\s*v?([^,]+?)[,)]", arr) { if(arr[1] ~ "^.+?[0-9]+\\.[0-9]+\\.[0-9]+(-.+)?$$") print arr[1]; exit; }')

ifeq ($(strip $(LAST_TAG)),)
	LAST_TAG = "1.0.0"
endif

# build as prerelease (dev) if not an exact tags
LATEST_STABLE_TAG := $(shell git -c versionsort.suffix="-rc" -c versionsort.suffix="-RC" tag -l "*.*.*" | sort -rV | awk '!/rc/' | head -n 1)
export IS_LATEST :=
ifeq ($(shell git describe --tags --exact-match --match $(LATEST_STABLE_TAG) >/dev/null 2>&1; echo $$?), 0)
	export IS_LATEST := true
else
	ifeq ($(prerelease),)
		prerelease := dev
	endif
endif

# Specify the release type manully, <major|minor|patch>, default release as last tag
# increase patch version when prerelease mode
release_as ?= $(LAST_TAG)
ifneq ($(prerelease),)
	release_as := patch
endif

get_version = \
	set -eu; \
	ver=$(LAST_TAG); \
	[ -n "$$ver" ] || exit 1; \
	release_as=$$(echo $(release_as) | sed "s/major/M/;s/minor/m/;s/patch/p/"); \
	ver=$$(echo "$$ver" | awk -v release_as=$$release_as 'BEGIN{FS=OFS="."} release_as~"^v?[0-9]+(\\.[0-9]+)*$$"{print gensub("^v","","g",release_as);exit} $$0~"(\\.[0-9]+)+$$"{ i=index("Mmp", release_as); if (i!=0) { $$i++; while (i<3) {$$(++i)=0} } print }'); \
	ver=$${ver:-$(LAST_TAG)}; \
	prerelease=$(prerelease); \
	if [ -n "$$prerelease" ]; then \
		ver="$${ver%%-$$prerelease}-$$prerelease"; \
	fi; \
	echo $$ver

release_tag := $(shell $(get_version))

.PHONY: build clean help test

# Parse Makefile and display the help
## > help - Show help
help:
	# Commands:
	@grep -E '^## > [a-zA-Z_-]+.*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "## > "}; {printf ">\033[36m%-1s\033[0m %s\n", $$1, $$2}'

build_args:=-ldflags "-w -s \
  -X main.appVersion=v$(release_tag) \
  -X main.gitCommit=$(GIT_COMMIT) \
  -X main.buildTime=$(BUILT)"

out: build

## > build [release_as=patch|minor|major] [prerelease=dev|rc|xxx] - build project for all support OSes
build:
	@echo "Compiling..."
	@echo "Build Version: $(release_tag)"
	$(GOX) $(build_args) -os "darwin linux windows" -arch="amd64 arm64" -output "out/$(NAME)-{{.OS}}-{{.Arch}}" ./

## > publish - publish to docker registry
publish: out
	# Publish calc-hash as docker ...
	@export BUILD_VERSION=$(release_tag) \
    && docker buildx bake \
      --set "main.args.BUILD_VERSION=$(release_tag)" \
      --set "main.args.BUILD_GIT_HEAD=$(GIT_COMMIT)" \
      --push

## > watch - watch source file for development build
watch:
	@watchman-make -p '**/*.go' -t build prerelease=dev

## > test [UPDATE_SNAPSHOTS=true] - run project tests
test:
	# Running tests...
	@go clean -testcache
	@go test $(OUR_PACKAGES) -cover -bench . -benchtime=5s

## > release - dry-run release
release:
	git release -t v -d

clean:
	rm -rf out
