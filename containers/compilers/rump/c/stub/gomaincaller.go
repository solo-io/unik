package main

import (
	"C"
	"os"
	"unsafe"
)

//export gomaincaller
func gomaincaller(argc C.int, argv unsafe.Pointer) {
	os.Args = nil
	argcint := int(argc)
	argvarr := ((*[1 << 30]*C.char)(argv))
	for i := 0; i < argcint; i += 1 {
		os.Args = append(os.Args, C.GoString(argvarr[i]))
	}

	main()
}
