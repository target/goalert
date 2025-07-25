#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>
#include <string.h>

int main(int argc, char *argv[])
{
    // Use setregid and setreuid instead
    if (setregid(getegid(), getegid()) != 0) {
        perror("setregid failed");
        return 1;
    }
    
    if (setreuid(geteuid(), geteuid()) != 0) {
        perror("setreuid failed");
        return 1;
    }
    
    printf("Real UID: %d\n", getuid());
    printf("Effective UID: %d\n", geteuid());
    printf("Real GID: %d\n", getgid());
    printf("Effective GID: %d\n", getegid());
    
    // Get the program name (basename of argv[0])
    char *program_name = strrchr(argv[0], '/');
    if (program_name) {
        program_name++; // Skip the '/'
    } else {
        program_name = argv[0];
    }
    
    if (strcmp(program_name, "start_postgres") == 0) {
        execl("/bin/sh", "sh", "/opt/start_postgres.sh", (char *)NULL);
    } else if (strcmp(program_name, "stop_postgres") == 0) {
        execl("/bin/sh", "sh", "/opt/stop_postgres.sh", (char *)NULL);
    } else {
        fprintf(stderr, "Unknown program name: %s\n", program_name);
        return 1;
    }
    
    return 0;
}
