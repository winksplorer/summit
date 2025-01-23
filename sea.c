#include <stdio.h>
#include <archive.h>
#include <archive_entry.h>
#include <stdlib.h>
#include <sys/stat.h>

extern const char _binary_summit_tar_gz_start[];
extern const char _binary_summit_tar_gz_end[];

void extract_mem(const void *data, size_t data_size, const char *output_dir) {
    struct archive *archive = archive_read_new(), *disk = archive_write_disk_new();
    struct archive_entry *entry;
    int r;

    archive_read_support_format_tar(archive);
    archive_read_support_filter_gzip(archive);
    archive_read_open_memory(archive, data, data_size);
    archive_write_disk_set_options(disk, ARCHIVE_EXTRACT_TIME | ARCHIVE_EXTRACT_PERM | ARCHIVE_EXTRACT_ACL | ARCHIVE_EXTRACT_FFLAGS);

    while ((r = archive_read_next_header(archive, &entry)) == ARCHIVE_OK) {
        char *path = NULL;
        asprintf(&path, "%s/%s", output_dir, archive_entry_pathname(entry));
        archive_entry_set_pathname(entry, path);

        if ((r = archive_write_header(disk, entry)) != ARCHIVE_OK) { perror("header write failed"); free(path); continue; }
        const void *buff;
        size_t size;
        la_int64_t offset;
        while ((r = archive_read_data_block(archive, &buff, &size, &offset)) == ARCHIVE_OK)
            if (archive_write_data_block(disk, buff, size, offset) != ARCHIVE_OK) { perror("data write failed"); break; }
        free(path);
        archive_write_finish_entry(disk);
    }

    if (r != ARCHIVE_EOF) perror("archive read failed");
    archive_read_free(archive);
    archive_write_free(disk);
}

int main() {
    printf("summit SEA (%s %s)\narchive size = %lu bytes\n", __DATE__, __TIME__, _binary_summit_tar_gz_end - _binary_summit_tar_gz_start);
    struct stat st;
    if (stat("/tmp/summit", &st) == 0 && S_ISDIR(st.st_mode)) rmdir("/tmp/summit");
    extract_mem(_binary_summit_tar_gz_start, _binary_summit_tar_gz_end - _binary_summit_tar_gz_start, "/tmp/summit");
    execv("/tmp/summit/server", (char *[]){"server", NULL});
    return 0;
}