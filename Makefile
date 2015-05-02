GOCMD="go"

PINGLIST_BINARY="pinglist"

.PHONY: build
build:
	$(GOCMD) build -race

.PHONY: clean
clean:
	rm -f $(PINGLIST_BINARY)

.PHONY: test
test:
	$(GOCMD) test -v
