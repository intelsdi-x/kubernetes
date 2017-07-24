#include <signal.h>
#include <stdio.h>
#include <sys/mman.h>
#include <stdlib.h>

int *pages = 0;
int pageNumber = 0;

void freePages() {
    fprintf(stderr, "\nFree memory\n");
    munmap(pages, pageNumber);
    fprintf(stderr, "========= EXITED =========\n");
    exit(0);
}

int main(int argc, char* argv[]) {
    if (argc != 2) {
        printf("Invalid number of arguments\n");
        return 3;
    }

    fprintf(stderr, "=== Reserving Hugepages ==\n\n");
    int numberOfPages = strtol(argv[1], (char **) NULL, 10);
    pageNumber = numberOfPages*2*1024*1024;

    pages = (int*) mmap(NULL, pageNumber, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS | MAP_HUGETLB, -1, 0);
    fprintf(stderr, "Huge pages has been reserved.\nNumber of hugepages: %d\nPointer: %p\nWaiting for interruption...\n", numberOfPages, pages);

    do {
        signal(SIGINT, freePages);
    } while (1);
    return 0;
}
