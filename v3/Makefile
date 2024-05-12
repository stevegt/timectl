.PHONY: build tag release

# The output binary name -- we derive this from the directory name
BINARY_NAME = $(shell basename $(PWD))

test:
	go test ./...

commit: test
	git add -A && grok commit | git commit -F- || echo "No changes to commit"

# Get major and minor from files.  
MAJOR = $(shell cat .version_major)
MINOR = $(shell cat .version_minor)

# Generate patch level based on number of commits.  The '=' instead
# of ':=' is important here, because we want to re-evaluate this each
# time it's used.  This is because the number of commits can change
# if we say 'make commit tag'.  (Due to the commit prerequisite, this
# will also happen if we say 'make tag'.)
PATCH = $(shell git rev-list HEAD --count)

# Put it all together
NEW_VERSION = v$(MAJOR).$(MINOR).$(PATCH)

tag: commit
	# throw error if major or minor version is not set
	[ -n "$(MAJOR)" ] || (echo "Set major version in .version_major"; exit 1)
	[ -n "$(MINOR)" ] || (echo "Set minor version in .version_minor"; exit 1)
	# throw error if uncommitted changes
	git diff-index --quiet HEAD -- || (echo "Commit or stash changes before tagging"; exit 1)
	git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"

# Extract version number from the latest git tag that looks like a version.  Falls back to v0.0.0 if no tag is found.
VERSION = $(shell git describe --tags --always --dirty --match="v[0-9]*\.[0-9]*\.[0-9]*" 2> /dev/null || echo "v0.0.0")

build: tag
	# XXX maybe use goreleaser here instead
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) .

push: tag
	git push --tags

release: push

