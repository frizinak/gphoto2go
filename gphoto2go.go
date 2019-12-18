package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

const (
	captureImage = C.GP_CAPTURE_IMAGE
	captureMovie = C.GP_CAPTURE_MOVIE
	captureSound = C.GP_CAPTURE_SOUND
)

// CameraEventType code
type CameraEventType int

const (
	eventUnknown   CameraEventType = C.GP_EVENT_UNKNOWN
	eventTimeout   CameraEventType = C.GP_EVENT_TIMEOUT
	eventFileAdded CameraEventType = C.GP_EVENT_FILE_ADDED
)

// CameraEvent struct
type CameraEvent struct {
	Type   CameraEventType
	Folder string
	File   string
}

// CameraFilePath struct
type CameraFilePath struct {
	Name   string
	Folder string
}

func cCameraEventToGoCameraEvent(voidPtr unsafe.Pointer, eventType C.CameraEventType) *CameraEvent {
	ce := new(CameraEvent)
	ce.Type = CameraEventType(eventType)

	if ce.Type == eventFileAdded {
		cameraFilePath := (*C.CameraFilePath)(voidPtr)
		ce.File = C.GoString((*C.char)(&cameraFilePath.name[0]))
		ce.Folder = C.GoString((*C.char)(&cameraFilePath.folder[0]))
	}

	return ce
}

func cameraListToMap(cameraList *C.CameraList) (map[string]string, int) {
	size := int(C.gp_list_count(cameraList))
	vals := make(map[string]string)

	if size < 0 {
		return vals, size
	}

	for i := 0; i < size; i++ {
		var cKey *C.char
		var cVal *C.char

		C.gp_list_get_name(cameraList, C.int(i), &cKey)
		C.gp_list_get_value(cameraList, C.int(i), &cVal)
		defer C.free(unsafe.Pointer(cKey))
		defer C.free(unsafe.Pointer(cVal))
		key := C.GoString(cKey)
		val := C.GoString(cVal)

		vals[key] = val
	}

	return vals, 0
}

func cameraResultToError(err C.int) error {
	if err != 0 {
		return fmt.Errorf(C.GoString(C.gp_result_as_string(err)))
	}
	return nil
}

// CameraResultToString func
func CameraResultToString(err C.int) string {
	return C.GoString(C.gp_result_as_string(err))
}
