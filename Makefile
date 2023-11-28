BUILD_DIR := .build
SVC_NAME := mongui

.PHONY: build run

all: build run

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(SVC_NAME) .

run:
	$(BUILD_DIR)/$(SVC_NAME)
