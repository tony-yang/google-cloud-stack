all: build-container

build-container:
	docker build -t google-cloud .