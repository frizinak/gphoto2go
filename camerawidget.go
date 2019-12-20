package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

type widgetValueType int

const (
	wvtString widgetValueType = iota
	wvtNum
	wvtDate
	wvtWeird
)

// WidgetTypeInfo struct
type WidgetTypeInfo struct {
	str   string
	vtype widgetValueType
	enum  C.CameraWidgetType
	desc  string
}

var widgetTypeTable = map[C.CameraWidgetType]WidgetTypeInfo{
	C.GP_WIDGET_WINDOW:  WidgetTypeInfo{"Window", wvtWeird, C.GP_WIDGET_WINDOW, "Window widget This is the toplevel configuration widget. It should likely contain multiple widget seciton entries"},
	C.GP_WIDGET_SECTION: WidgetTypeInfo{"Section", wvtWeird, C.GP_WIDGET_SECTION, "Section widget (think Tab)"},
	C.GP_WIDGET_TEXT:    WidgetTypeInfo{"Text", wvtString, C.GP_WIDGET_TEXT, "Text widget"},
	C.GP_WIDGET_RANGE:   WidgetTypeInfo{"Range", wvtNum, C.GP_WIDGET_RANGE, "Slider widget"},
	C.GP_WIDGET_TOGGLE:  WidgetTypeInfo{"Toggle", wvtNum, C.GP_WIDGET_TOGGLE, "Toggle widget (think check box)"},
	C.GP_WIDGET_RADIO:   WidgetTypeInfo{"Radio", wvtString, C.GP_WIDGET_RADIO, "Radio button widget"},
	C.GP_WIDGET_MENU:    WidgetTypeInfo{"Menu", wvtNum, C.GP_WIDGET_MENU, "Menu widget (same as RADIO)"},
	C.GP_WIDGET_BUTTON:  WidgetTypeInfo{"Button", wvtNum, C.GP_WIDGET_BUTTON, "Button press widget"},
	C.GP_WIDGET_DATE:    WidgetTypeInfo{"Date", wvtDate, C.GP_WIDGET_DATE, "Date entering widget"},
}

// Str get the widget's type as a string
func (wti *WidgetTypeInfo) Str() string {
	return wti.str
}

// CameraWidget struct
type CameraWidget struct {
	widget *C.CameraWidget
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

// Free func
func (w *CameraWidget) Free() {
	if err := cameraResultToError(C.gp_widget_free(w.widget)); err != nil {
		fmt.Printf("WARNING: error on C.gp_widget_free: %v\n", err)
	}
}

//
// getters and setters
//

// Name func
func (w *CameraWidget) Name() (string, error) {
	var _name *C.char
	defer C.free(unsafe.Pointer(_name))

	if err := cameraResultToError(C.gp_widget_get_name(w.widget, &_name)); err != nil {
		return "", err
	}

	return ToString(_name), nil
}

// Value func
func (w *CameraWidget) Value() (interface{}, error) {
	var err error
	wti, err := w.Type()
	if err != nil {
		return nil, err
	}

	switch wti.vtype {
	case wvtString:
		var val *C.char
		defer C.free(unsafe.Pointer(val)) // copied by ToString()
		if err := cameraResultToError(C.gp_widget_get_value(w.widget, unsafe.Pointer(&val))); err != nil {
			return nil, err
		}
		return ToString(val), nil
	case wvtNum:
		var val = new(int)
		if err := cameraResultToError(C.gp_widget_get_value(w.widget, unsafe.Pointer(val))); err != nil {
			return nil, err
		}
		return *val, nil
	case wvtDate:
		var val = new(int64)
		if err := cameraResultToError(C.gp_widget_get_value(w.widget, unsafe.Pointer(val))); err != nil {
			return nil, err
		}
		return time.Unix(*val, 0), nil
	case wvtWeird:
	default:
		return "", nil
	}

	return nil, err
}

// Parent func
func (w *CameraWidget) Parent() (*CameraWidget, error) {
	_parent := new(CameraWidget)

	if err := cameraResultToError(C.gp_widget_get_parent(w.widget, &_parent.widget)); err != nil {
		return nil, err
	}

	return _parent, nil
}

// Label func
func (w *CameraWidget) Label() (string, error) {
	var _label *C.char
	defer C.free(unsafe.Pointer(_label))

	if err := cameraResultToError(C.gp_widget_get_label(w.widget, &_label)); err != nil {
		return "", err
	}

	return ToString(_label), nil
}

// Type func
func (w *CameraWidget) Type() (*WidgetTypeInfo, error) {
	var _type C.CameraWidgetType

	if err := cameraResultToError(C.gp_widget_get_type(w.widget, &_type)); err != nil {
		return new(WidgetTypeInfo), err
	}

	widgetTypeInfo := widgetTypeTable[_type]
	return &widgetTypeInfo, nil
}

// Readonly func
func (w *CameraWidget) Readonly() (bool, error) {
	var _ro C.int

	if err := cameraResultToError(C.gp_widget_get_readonly(w.widget, &_ro)); err != nil {
		return false, err
	}

	return _ro == 1, nil
}

// Child func
func (w *CameraWidget) Child(name string) (*CameraWidget, error) {
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

// ValueType func
func (w *CameraWidget) ValueType() (string, error) {
	wti, err := w.Type()
	if err != nil {
		return "", err
	}
	switch wti.vtype {
	case wvtString:
		return "string", nil
	case wvtNum:
		return "int", nil
	case wvtDate:
		return "date", nil
	default:
		return "weird", nil
	}
}
