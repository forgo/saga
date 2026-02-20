.PHONY: dev test lint fmt check build clean

dev:
	deno task dev

test:
	deno test --allow-all

lint:
	deno lint

fmt:
	deno fmt

check: lint test

build:
	deno task build

clean:
	rm -rf _fresh
