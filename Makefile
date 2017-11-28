all: lint test build

build: _output/nodeport-exposer

_output/nodeport-exposer:
		mkdir -p _output
		go build -v -ldflags '-w -extldflags -static' -o _output/nodeport-exposer ./cmd/nodeport-exposer

test:
	go test -v ./...

lint:
	{ \
	set -e ;\
	PACKAGES=$$(go list ./...) ;\
	echo "go vet" ;\
	go vet $$PACKAGES ;\
	echo "golint" ;\
	golint $$PACKAGES ;\
	echo "errcheck -blank" ;\
	errcheck -blank $$PACKAGES ;\
	echo "varcheck" ;\
	varcheck $$PACKAGES ;\
	echo "structcheck" ;\
	structcheck $$PACKAGES ;\
	echo "gosimple" ;\
	gosimple $$PACKAGES ;\
	echo "unused" ;\
	unused $$PACKAGES ;\
	GOFILES=$$(find . -type f -name '*.go' -not -path "./vendor/*") ;\
	echo "misspell -error -locale US" ;\
	misspell -error -locale US $$GOFILES ;\
	}
