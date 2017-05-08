ifeq ($(IMAGE_NAME),)
IMAGE_NAME:=cgroup-hugepage-isolator
endif

all: binary docker

binary:
	go build -v -o hugepage-isolator

docker: binary
	docker build -t ${IMAGE_NAME}:latest .

clean:
	rm -rf hugepage-isolator > /dev/null 2>&1
