package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
	char *bucket_pid;
	bucket_pid = getenv("bucket_pid");
	if (bucket_pid) {
		//fprintf(stdout, "got bucket_pid=%s\n", bucket_pid);
	} else {
		//fprintf(stdout, "missing bucket_pid env skip nsenter");
		return;
	}
	char *bucket_cmd;
	bucket_cmd = getenv("bucket_cmd");
	if (bucket_cmd) {
		//fprintf(stdout, "got bucket_cmd=%s\n", bucket_cmd);
	} else {
		//fprintf(stdout, "missing bucket_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", bucket_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(bucket_cmd);
	exit(0);
	return;
}
*/
import "C"
