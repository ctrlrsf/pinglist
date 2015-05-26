GOCMD="go"

PINGLIST_BINARY="pinglist"

.PHONY: all
all: build build-js

.PHONY: build
build:
	$(GOCMD) build -race

.PHONY: build-js
build-js:
	rm -rf public
	cp -a static public
	jsx static/ public/

.PHONY: clean
clean:
	rm -rf public
	rm -f $(PINGLIST_BINARY)

.PHONY: test
test:
	$(GOCMD) test -v
