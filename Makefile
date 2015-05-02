GOCMD="go"

.PHONY: build
build:
	$(GOCMD) build -race


.PHONY: test
test:
	$(GOCMD) test -v
