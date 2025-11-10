BIN ?= dist/rss-urls

add_upload: add
	cp rss.xml archive/rss-$(shell date +%Y-%m-%d-%H%M).xml
	make upload

restore:
	cp $(shell ls -t archive/rss-*.xml | head -n 1) rss.xml

build:
	go build -v -o $(BIN) .

add: build
	$(BIN) $(URL)

upload:
	aws s3 cp rss.xml s3://$(BUCKET_NAME)/rss.xml --content-type application/rss+xml

TEST ?= ./...

lint:
	golangci-lint run $(TEST)
lint.fix:
	golangci-lint run --fix $(TEST)