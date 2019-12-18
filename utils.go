package gphoto2go

// #include <stdlib.h>
import "C"

// ToString func
func ToString(charPtr *C.char) (string) {
	model := C.GoString((*C.char)(charPtr))
	return model
}