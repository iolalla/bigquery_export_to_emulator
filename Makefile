export CGO_ENABLED := 0
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## "Building a valid production binary"
	echo "Building a valid production binary"
	go build -o be_exp

bu: ## "Looks like we have a binary ready to debug"
	echo "Looks like we have a binary ready to debug"
	go build  -gcflags="all=-N -l" -o be_exp_dbg

all: build ## "Executes Build as there are not tests or deploy process"
