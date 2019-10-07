package parser

// #cgo LDFLAGS: -Ltarget/release -lflux_parser
// #include <stdlib.h>
// void flux_parse_json(const char*);
import "C"

import (
	"unsafe"
)

func Parse(input string) {
	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	C.flux_parse_json(cstr)
}
