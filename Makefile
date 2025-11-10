BIN ?= dist/rss-urls

add_upload:
	$(BIN) $(URL)
	cp rss.xml archive/rss-$(shell date +%Y-%m-%d-%H%M).xml
	make upload

build:
	go build -v -o $(BIN) .

add: build

upload:
	aws s3 cp rss.xml s3://$(BUCKET_NAME)/rss.xml --content-type application/rss+xml

TEST ?= ./...

lint:
	golangci-lint run $(TEST)
lint.fix:
	golangci-lint run --fix $(TEST)