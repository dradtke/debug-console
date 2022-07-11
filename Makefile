build:
	mkdir -p bin
	go build -o bin -tags nvim ./cmd/debug-console/
	bin/debug-console nvim -manifest debug-console -location misc/nvim/debug-console.vim

.PHONY: build
