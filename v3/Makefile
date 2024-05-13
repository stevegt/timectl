# Workflow:  Develop on main, release on branch.  Release tags go only
# on branches.  

# To push changes without a release:  make sync

# To merge from main, tag, and push:  make release

# To push a fix to a release branch without merging from main:  make fix

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

# Branch we'll do the release on
BRANCH_TARGET = v$(MAJOR)
# Extract version number from the latest git tag that looks like a version.  Falls back to v0.0.0 if no tag is found.
VERSION = $(shell git describe --tags --always --dirty --match="v[0-9]*\.[0-9]*\.[0-9]*" 2> /dev/null || echo "v0.0.0")

.PHONY: release fix test ck-main ck-branch ck-commit commit merge tag build push

# The output binary name -- we derive this from the directory name
BINARY_NAME = $(shell basename $(PWD))

sync: ck-main test commit push

release: ck-main test commit merge test tag push

fix: ck-branch test commit tag push

test:
	go test ./...

ck-main:
	# fail if we're not on main
	git branch --show-current | egrep -q "^main$$" || (echo "Not on main branch"; exit 1)

ck-branch:
	# fail if we're not on the target branch
	git branch --show-current | egrep -q "^$(BRANCH_TARGET)$$" || (echo "Not on $(BRANCH_TARGET) branch"; exit 1)

ck-commit:
	# throw error if uncommitted changes
	git diff-index --quiet HEAD -- || (echo "Commit or stash changes before tagging"; exit 1)

commit: ck-main
	git add -A && grok commit | git commit -F- || echo "No changes to commit"

merge: ck-commit
	# switch to branch
	git switch $(BRANCH_TARGET)
	git merge --commit main 

tag: ck-branch ck-commit test
	# throw error if major or minor version is not set
	[ -n "$(MAJOR)" ] || (echo "Set major version in .version_major"; exit 1)
	[ -n "$(MINOR)" ] || (echo "Set minor version in .version_minor"; exit 1)
	git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"

build: 
	# XXX maybe use goreleaser here instead
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) .

push: 
	git push --tags


