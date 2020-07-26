package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
import "C"
import "fmt"

const (
	ErrModelNotFound = C.GP_ERROR_MODEL_NOT_FOUND
)

const (
	errUnknown = "libgphoto2: unknown error"
)

type Error struct {
	code    int
	message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("libgphoto2: [%d] %s", e.code, e.message)
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) IsModelNotFound() bool {
	return e.code == ErrModelNotFound
}

func cameraResultToError(code C.int) error {
	if code == 0 {
		return nil
	}

	str := C.GoString(C.gp_result_as_string(code))
	if str == "" {
		str = errUnknown
	}

	return &Error{
		code:    int(code),
		message: str,
	}
}

// CameraResultToString func
func CameraResultToString(err C.int) string {
	return C.GoString(C.gp_result_as_string(err))
}
