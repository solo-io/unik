
int kludge_argc = 1;
char *kludge_argv[] = { "foo", 0 };

int main() {
 	// rump_pub_lwproc_releaselwp(); /* XXX */
	gomaincaller();
	return 0;
}
