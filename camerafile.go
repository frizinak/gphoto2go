package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"

// CameraFile struct
type CameraFile struct {
	file  *C.CameraFile
	cSize C.ulong
	buf   *C.char
}
