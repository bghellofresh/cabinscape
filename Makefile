NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: all clean deps build

all: clean deps build

clean:
	@printf "$(OK_COLOR)==> Cleaning project$(NO_COLOR)\n"
	@rm -rf dist

deps:
	@printf "$(OK_COLOR)==> Installing deps using gopkg.toml$(NO_COLOR)\n"
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure

build:
	@printf "$(OK_COLOR)==> Building binary$(NO_COLOR)\n"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o ./dist/airbnb ./
