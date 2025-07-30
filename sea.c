#include <stdio.h>
#include <archive.h>
#include <archive_entry.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/stat.h>

#define print_archive_error_exit(msg, archive) return fprintf(stderr, "%s: %s\n", msg, archive_error_string(archive)), -1
extern const char _binary_summit_tar_xz_start[], _binary_summit_tar_xz_end[];

int extract_mem(const void *data, size_t data_size, const char *output_dir) {
    // init libarchive + some variables
    struct archive* archive = archive_read_new(), * disk = archive_write_disk_new(); // archive = summit.tar.gz, disk = /tmp/summit
    if (!archive) return perror("archive_read_new"), -1;
    if (!disk) return perror("archive_write_disk_new"), archive_read_free(archive), -1; // i'll regret this later
    struct archive_entry* entry; int r; const void* buff; size_t size; la_int64_t offset;

    // libarchive options
    if (archive_read_support_format_tar(archive))                      print_archive_error_exit("archive_read_support_format_tar", archive);
    if (archive_read_support_filter_xz(archive))                       print_archive_error_exit("archive_read_support_filter_xz", archive);
    if (archive_read_open_memory(archive, data, data_size)) print_archive_error_exit("archive_read_open_memory", archive);
    if (archive_write_disk_set_options(disk,
        ARCHIVE_EXTRACT_TIME | ARCHIVE_EXTRACT_PERM | ARCHIVE_EXTRACT_SECURE_NODOTDOT
       | ARCHIVE_EXTRACT_ACL | ARCHIVE_EXTRACT_FFLAGS))                 print_archive_error_exit("archive_write_disk_set_options", archive);

    // go through the archive file by file
    while (!(r = archive_read_next_header(archive, &entry))) {
        // construct & set path
        char* path;
        if (asprintf(&path, "%s/%s", output_dir, archive_entry_pathname(entry)) == -1) {
            perror("asprintf");
            break;
        }
        archive_entry_set_pathname(entry, path);

        // write the header
        if ((r = archive_write_header(disk, entry))) { 
            fprintf(stderr, "header write failed at %s: %s\n", archive_entry_pathname(entry), archive_error_string(disk));
            free(path);
            continue;
        }

        // go through the file block by block, and write each block
        while (!(r = archive_read_data_block(archive, &buff, &size, &offset))) 
            if (archive_write_data_block(disk, buff, size, offset)) {
                fprintf(stderr, "data write failed at %s: %s\n", archive_entry_pathname(entry), archive_error_string(disk));
                break;
            }

        free(path);
        archive_write_finish_entry(disk);
    }

    // cleanup
    if (r != ARCHIVE_EOF) fprintf(stderr, "archive read failed: %s\n", archive_error_string(archive));
    if (archive_read_free(archive)) fprintf(stderr, "archive_read_free: %s\n", archive_error_string(archive));
    if (archive_write_free(disk)) fprintf(stderr, "archive_write_free: %s\n", archive_error_string(archive));
    return r == ARCHIVE_EOF ? 0 : -1;
}

int main(int argc, char* argv[]) {
    // root check + printing basic info
    if (geteuid() != 0) return fprintf(stderr, "summit requires root permissions to work correctly.\n"), 1;
    printf("summit SEA (%s %s).\nsummit and its SEA are licensed under Apache-2.0 and include third-party components under MIT and OFL licenses. copyright (c) 2025 winksplorer et al.\narchive size = %.2f MB\n",
        __DATE__, __TIME__, (double)(_binary_summit_tar_xz_end - _binary_summit_tar_xz_start)/1000000);

    // extract the embedded summit.tar.xz into /tmp/summit
    if (extract_mem(_binary_summit_tar_xz_start, _binary_summit_tar_xz_end - _binary_summit_tar_xz_start, "/tmp/summit") == -1)
        return fprintf(stderr, "archive extract failed\n"), 1;

    // pass through arguments
    argv[0] = "/tmp/summit/summit-server";
    execv("/tmp/summit/summit-server", argv);
    return perror("execv"), 1;
}