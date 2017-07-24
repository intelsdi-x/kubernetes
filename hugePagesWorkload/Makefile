all:
	$(MAKE) build
	$(MAKE) pack

build:
	gcc -o ./hugepagesAllocator hugepagesAllocator.c

pack:
	docker build -t hugepagesallocator:latest .

clean:
	rm ./hugepagesAllocator
