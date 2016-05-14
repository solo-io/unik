package main

import (
	"flag"
	"fmt"
	"strings"
	"path/filepath"
	"log"
	"io/ioutil"
)

func main() {
	argString := flag.String("args", "", "arguments as a single string")
	outFile := flag.String("o", "", "output file to write to")
	flag.Parse()
	args := []string{"\"program.bin\""}
	for _, arg := range strings.Split(*argString, " ") {
		args = append(args,
		fmt.Sprintf("\"%s\"", arg))
	}
	cmainstub := fmt.Sprintf(`
int kludge_argc = %v;
char *kludge_argv[] = { %s, 0 };

int main() {
 	rump_pub_lwproc_releaselwp(); /* XXX */
	gomaincaller();
}

	`, len(args), strings.Join(args, ", "))
	stubFile, err := filepath.Abs(*outFile)
	if err != nil {
		log.Fatalf("failed with err %v", err)
	}
	if err := ioutil.WriteFile(stubFile, []byte(cmainstub), 0644); err != nil {
		log.Fatalf("failed with err %v", err)
	}
	log.Printf("wrote to %s:\n%s", stubFile, cmainstub)
}