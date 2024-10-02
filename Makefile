PORT ?= 8080
PROG_NAME ?= pseudo-web-app
BIN_DIR ?= bin
LOGLEVEL ?= info
STATE_DIR ?= state

STATE_COUNT_FILE := $(STATE_DIR)/count

# go build -a -installsuffix cgo -o /install/bin/pseudo-web-app
DOCKER ?= $(shell for docker_prog in podman docker; do which $${docker_prog} 2> /dev/null && exit 0; done)
DOCKER_IMAGE = $(shell basename $(shell pwd))

$(warning DOCKER: $(DOCKER))
$(warning DOCKER_IMAGE: ${DOCKER_IMAGE})

.PHONY: help build serve state clean
build: $(BIN_DIR) $(BIN_DIR)/$(PROG_NAME)

SOURCE_FILES=$(wildcard */*.go */*/*.go)

$(BIN_DIR):
	@[ -e "$(BIN_DIR)" ] || mkdir -pv "$(BIN_DIR)"

$(BIN_DIR)/$(PROG_NAME): $(SOURCE_FILES)
	go build -o $(BIN_DIR)/$(PROG_NAME) ./app/main.go

serve: $(STATE_COUNT_FILE)
	@echo "Starting app - press CTRL-C to stop..."
	PORT=$(PORT) LOGLEVEL=$(LOGLEVEL) $(BIN_DIR)/$(PROG_NAME)

state: $(STATE_COUNT_FILE)

clean:
	@([ -e "$(BIN_DIR)/$(PROG_NAME)" ] && rm -v "$(BIN_DIR)/$(PROG_NAME)") || true
	@([ -e "$(BIN_DIR)" ] && rmdir -v $(BIN_DIR)) || true

superclean: clean
	@([ -e "$(STATE_COUNT_FILE)" ] && rm -v "$(STATE_COUNT_FILE)") || true

$(STATE_COUNT_FILE):
	@echo "Creating missing state file '$@'..."
	@([ -d "$(STATE_DIR)" ] || mkdir -v "$(STATE_DIR)") || true
	@([ -e "$@" ] ||	printf "0" > "$@") || true

help:
	@echo "make build	- build the app"
	@echo "make serve	- run the app"
	@echo "make clean	- clean the executable
	@echo "make superclean	- like 'make clean', but also remove the preserved state"

ifeq ($(DOCKER),)

  $(Warning Could not find docker nor podman in the PATH, you will not be able to work with containers)

else

  .PHONY:  docker-image docker-run podman-init podman-start

  docker-image:
	${DOCKER) build -t ${DOCKER_IMAGE} .

  docker-run:
	$(DOCKER) run --env-file .env -p $(PORT):$(PORT) ${DOCKER_IMAGE}

  podman-init:
	podman machine init

  podman-start:
	podman machine start

endif
