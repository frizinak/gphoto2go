package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
import "C"
import "fmt"

const (
	Err                   = C.GP_ERROR
	ErrBadParameters      = C.GP_ERROR_BAD_PARAMETERS
	ErrNoMemory           = C.GP_ERROR_NO_MEMORY
	ErrLibrary            = C.GP_ERROR_LIBRARY
	ErrUnknownPort        = C.GP_ERROR_UNKNOWN_PORT
	ErrNotSupported       = C.GP_ERROR_NOT_SUPPORTED
	ErrIO                 = C.GP_ERROR_IO
	ErrFixedLimitExceeded = C.GP_ERROR_FIXED_LIMIT_EXCEEDED
	ErrTimeout            = C.GP_ERROR_TIMEOUT
	ErrIOSupportedSerial  = C.GP_ERROR_IO_SUPPORTED_SERIAL
	ErrIOSupportedUSB     = C.GP_ERROR_IO_SUPPORTED_USB
	ErrIOInit             = C.GP_ERROR_IO_INIT
	ErrIORead             = C.GP_ERROR_IO_READ
	ErrIOWrite            = C.GP_ERROR_IO_WRITE
	ErrIOUpdate           = C.GP_ERROR_IO_UPDATE
	ErrIOSerialSpeed      = C.GP_ERROR_IO_SERIAL_SPEED
	ErrIOUSBClearHalt     = C.GP_ERROR_IO_USB_CLEAR_HALT
	ErrIOUSBFind          = C.GP_ERROR_IO_USB_FIND
	ErrIOUSBClaim         = C.GP_ERROR_IO_USB_CLAIM
	ErrIOLock             = C.GP_ERROR_IO_LOCK
	ErrHal                = C.GP_ERROR_HAL
	ErrCorruptedData      = C.GP_ERROR_CORRUPTED_DATA
	ErrFileExists         = C.GP_ERROR_FILE_EXISTS
	ErrModelNotFound      = C.GP_ERROR_MODEL_NOT_FOUND
	ErrDirectoryNotFound  = C.GP_ERROR_DIRECTORY_NOT_FOUND
	ErrFileNotFound       = C.GP_ERROR_FILE_NOT_FOUND
	ErrDirectoryExists    = C.GP_ERROR_DIRECTORY_EXISTS
	ErrCameraBusy         = C.GP_ERROR_CAMERA_BUSY
	ErrPathNotAbsolute    = C.GP_ERROR_PATH_NOT_ABSOLUTE
	ErrCancel             = C.GP_ERROR_CANCEL
	ErrCameraError        = C.GP_ERROR_CAMERA_ERROR
	ErrOsFailure          = C.GP_ERROR_OS_FAILURE
	ErrNoSpace            = C.GP_ERROR_NO_SPACE
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

func (e *Error) Is(i int) bool {
	return e.code == i
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
