
int kludge_argc = 1;
char *kludge_argv[] = { "foo", 0, 0};

extern unsigned int rumpns_kludge_dns_addrs[16];
extern unsigned int rumpns_kludge_dns_addr_count;

int main(int argc, char *argv[]) {
    printf("got this: %d\n", rumpns_kludge_dns_addr_count);
	gomaincaller(rumpns_kludge_dns_addr_count * 4, rumpns_kludge_dns_addrs, argc, argv);
	return 0;
}
