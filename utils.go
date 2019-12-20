package gphoto2go

// #include <stdlib.h>
import "C"

// ToString func
func ToString(charPtr *C.char) string {
	gostring := C.GoString((*C.char)(charPtr))
	return gostring
}
