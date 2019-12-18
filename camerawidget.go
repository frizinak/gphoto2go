package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// CameraWidget struct
type CameraWidget struct {
	widget *C.CameraWidget
}

// Name func
func (w *CameraWidget) Name() (string, error) {
	var _name *C.char
	defer C.free(unsafe.Pointer(_name))

	if err := cameraResultToError(C.gp_widget_get_name(w.widget, &_name)); err != nil {
		return "", err
	}

	return ToString(_name), nil
}

// Free func
func (w *CameraWidget) Free() {
	if err := cameraResultToError(C.gp_widget_free(w.widget)); err != nil {
		fmt.Printf("WARNING: error on C.gp_widget_free: %v\n", err)
	}
}

// GetChildrenByName func
func (w *CameraWidget) GetChildrenByName(name string) (*CameraWidget, error) {
	var child *C.CameraWidget

	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))

	err := cameraResultToError(C.gp_widget_get_child_by_name(w.widget, n, &child))
	if err != nil {
		C.free(unsafe.Pointer(child))
		return nil, fmt.Errorf("error on C.gp_widget_get_child_by_name(%s): %v", name, err)
	}

	return &CameraWidget{child}, nil
}

// SetValue func
func (w *CameraWidget) SetValue(v interface{}) error {
	switch v.(type) {
	case string:
		cstr := C.CString(v.(string))
		defer C.free(unsafe.Pointer(cstr))

		if err := cameraResultToError(C.gp_widget_set_value(w.widget, unsafe.Pointer(cstr))); err != nil {
			return err
		}
	default:
		if err := cameraResultToError(C.gp_widget_set_value(w.widget, unsafe.Pointer(&v))); err != nil {
			return err
		}
	}

	return nil
}
