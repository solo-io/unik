
int kludge_argc = 1;
char *kludge_argv[] = { "foo", 0, 0};

extern unsigned int kludge_dns_addr;

int main(int argc, char *argv[]) {
	gomaincaller(kludge_dns_addr, argc, argv);
	return 0;
}
