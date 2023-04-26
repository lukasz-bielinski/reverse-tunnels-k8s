.PHONY: build push

IMAGE_NAME := lukaszbielinski/chisel-register-api

build:
		docker build . -t $(IMAGE_NAME)

push: build
		docker push $(IMAGE_NAME)

