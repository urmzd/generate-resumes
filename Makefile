# Makefile for Generate-Resumes project

# Variables
BASE_IMAGE_NAME := generate-resumes-base
ORGANIZATION := urmzd
VERSION ?= latest

# Paths
OUTPUTS_DIR := outputs
INPUTS_DIR := inputs
EXAMPLES_DIR := examples

# Docker image names
BASE_IMAGE_TAG := $(ORGANIZATION)/$(BASE_IMAGE_NAME):$(VERSION)
APP_IMAGE_TAG := $(ORGANIZATION)/generate-resumes

# Build the base docker image
build-base:
	@echo "Building base Docker image..."
	docker build -t $(BASE_IMAGE_NAME) -f base.Dockerfile .

# Tag and push the base image to Docker Hub
push-base:
	@echo "Tagging and pushing base image to Docker Hub..."
	docker tag $(BASE_IMAGE_NAME) $(BASE_IMAGE_TAG)
	docker push $(BASE_IMAGE_TAG)

# Build the application docker image
build:
	@echo "Building application Docker image..."
	docker build -t $(APP_IMAGE_TAG) .

# Initialize project directories and copy examples to inputs
init:
	@echo "Initializing directories and copying example files..."
	mkdir -p $(OUTPUTS_DIR) $(INPUTS_DIR)
	cp $(EXAMPLES_DIR)/* $(INPUTS_DIR)/

# Run the application in Docker
run:
	@echo "Running application in Docker..."
	docker run -v "$(shell pwd)/$(OUTPUTS_DIR):/outputs" -v "$(shell pwd)/$(INPUTS_DIR):/inputs" $(APP_IMAGE_TAG) /inputs/$(FILENAME) -o /outputs

# Clean outputs and inputs directories
clean:
	@echo "Cleaning outputs and inputs directories..."
	rm -rf $(OUTPUTS_DIR) $(INPUTS_DIR)

# Run examples
run-examples:
	@echo "Running examples..."
	make init
	cp -r $(EXAMPLES_DIR)/ $(INPUTS_DIR)
	make run FILENAME=example.yml
