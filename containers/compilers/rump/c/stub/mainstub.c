
int kludge_argc = 1;
char *kludge_argv[] = { "foo", 0 };

int main(int argc, char *argv[]) {
	gomaincaller(argc, argv);
	return 0;
}