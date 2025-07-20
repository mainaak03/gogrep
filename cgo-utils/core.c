#define _GNU_SOURCE // Adding this to enable use of memrchr, it is SIMD optimised and probably better than the while loop I would write otherwise

#define RED   "\033[1;31m"
#define RESET_COLOR "\033[0m"

#include <stdio.h>
#include <string.h>
#include <dirent.h>
#include <stdbool.h>
#include <unistd.h>
#include <stddef.h>
#include <stdlib.h>
#include <fcntl.h>
#include <fnmatch.h>
#include <sys/stat.h>
#include <sys/mman.h>

#include <rure.h>

extern void walkAndMatch(const char *path, const char *pattern, const bool recursive);
void processFile(const char *filepath);
bool isDir(const char *path);
void walkDir(const char *path, const char *pattern, const bool recursive);

// should compile regex only once as it takes some ms and is resource heavy lol
rure *re;

void walkAndMatch(const char *path, const char *pattern, const bool recursive) {
    if (!re) re = rure_compile_must(pattern);

    if (isDir(path)) {
        walkDir(path, pattern, recursive);
    } else {
        processFile(path);
    }
    
    rure_free(re);
}

void processFile(const char *filepath) {
    int fd = open(filepath, O_RDONLY);
    if (fd == -1) {
        perror("failed to open file");
        exit(1);
    }
    struct stat status;
    if (fstat(fd, &status) == -1) {
        perror("failed to stat file");
        exit(1);
    }
    if (status.st_size == 0) {
        return;
    }

    size_t fileLen = (size_t)status.st_size;
    const uint8_t *mmapFile = (const uint8_t *)mmap(NULL, status.st_size, PROT_READ, MAP_PRIVATE, fd, 0);
    close(fd);
    if (mmapFile == MAP_FAILED) {
        perror("could not mmap file");
        exit(1);
    }

    rure_iter *it = rure_iter_new(re);
    rure_match match;

    bool isBinary = memchr(mmapFile, '\0', (fileLen < 1024) ? fileLen : 1024) != NULL;

    if (!isBinary) {
        const uint8_t *start = mmapFile;
        const uint8_t *fileEnd = mmapFile + fileLen;
        const uint8_t *newline;

        while (start < fileEnd) {
            newline = memchr(start, '\n', fileEnd - start);
            const uint8_t *lineEnd = newline ? newline : fileEnd;
            size_t lineLen = lineEnd - start;

            if (rure_find(re, start, lineLen, 0, &match)) {
                fwrite(filepath, 1, strlen(filepath), stdout);
                fwrite(": ", 1, 2, stdout);
                fwrite(start, 1, match.start, stdout);
                fwrite(RED, 1, strlen(RED), stdout);
                fwrite(start + match.start, 1, match.end - match.start, stdout);
                fwrite(RESET_COLOR, 1, strlen(RESET_COLOR), stdout);
                fwrite(start + match.end, 1, lineLen - match.end, stdout);
                fwrite("\n", 1, 1, stdout);
            }
            start = newline ? newline + 1 : fileEnd;
        }
    }

    munmap((char*)mmapFile, fileLen);
    rure_iter_free(it);
}

bool isDir(const char *path) {
    struct stat st;
    return stat(path, &st) == 0 && S_ISDIR(st.st_mode);
}

void walkDir(const char *path, const char *pattern, const bool recursive) {
    DIR *dir = opendir(path);
    if (!dir) {
        perror("opendir");
        return;
    }

    struct dirent *entry;
    char fullpath[4096];

    while ((entry = readdir(dir)) != NULL) {
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0)
            continue;

        snprintf(fullpath, sizeof(fullpath), "%s/%s", path, entry->d_name);

        if (isDir(fullpath)) {
            if (recursive) walkDir(fullpath, pattern, recursive);
        } else {
            processFile(fullpath);
        }
    }

    closedir(dir);
}
