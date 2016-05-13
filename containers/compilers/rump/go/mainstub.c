
int main() {
 	// rump_pub_lwproc_releaselwp(); 
	 /* see ultimate kludge: https://github.com/deferpanic/gorump/commit/ab63d5a1389aba2a588dbc2a956b1ea97e27dc53 
	 do not use this work around if you expect your program to exit, as it will panic*/
	gomaincaller();
	return 0;
}
