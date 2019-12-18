package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"io"
	"reflect"
	"unsafe"
)

// CameraFile struct
type CameraFile struct {
	file *C.CameraFile
}

// Need to find a good buffer size
// For now, let's try 1MB
const fileReaderBufferSize = 1 * 1024 * 1024

type cameraFileReader struct {
	camera   *Camera
	folder   string
	fileName string
	fullSize uint64
	offset   uint64
	closed   bool

	cCameraFile *C.CameraFile
	cBuffer     *C.char

	buffer [fileReaderBufferSize]byte
}

func (cfr *cameraFileReader) Read(p []byte) (int, error) {
	if cfr.closed {
		return 0, io.ErrClosedPipe
	}

	n := uint64(len(p))

	if n == 0 {
		return 0, nil
	}

	bufLen := uint64(len(cfr.buffer))
	remaining := cfr.fullSize - cfr.offset

	toRead := bufLen
	if toRead > remaining {
		toRead = remaining
	}

	if toRead > n {
		toRead = n
	}

	// From: https://code.google.com/p/go-wiki/wiki/cgo
	// Turning C arrays into Go slices
	sliceHeader := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cfr.cBuffer)),
		Len:  int(cfr.fullSize),
		Cap:  int(cfr.fullSize),
	}
	goSlice := *(*[]C.char)(unsafe.Pointer(&sliceHeader))

	for i := uint64(0); i < toRead; i++ {
		p[i] = byte(goSlice[cfr.offset+i])
	}

	cfr.offset += toRead

	if cfr.offset < cfr.fullSize {
		return int(toRead), nil
	}
	return int(toRead), io.EOF
}

func (cfr *cameraFileReader) Close() error {
	if !cfr.closed {
		// If I understand correctly, freeing the CameraFile will also free the data buffer (ie. cfr.cBuffer)
		C.gp_file_free(cfr.cCameraFile)
		cfr.closed = true
	}
	return nil
}
