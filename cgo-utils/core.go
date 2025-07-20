package cgoutils

//
// #cgo LDFLAGS: -lrure
// #include <stdint.h>
// #include <stdbool.h>
// #include <stdlib.h>
//
//void walkAndMatch(const char *path, const char *pattern, bool recursive);
import "C"
import ("unsafe")

func WalkAndMatch(path, pattern string, recursive bool) {
	cPath := C.CString(path)
	cPattern := (C.CString(pattern))
	defer C.free(unsafe.Pointer(cPath))
	defer C.free(unsafe.Pointer(cPattern))

	C.walkAndMatch(cPath, cPattern, C.bool(recursive))
}