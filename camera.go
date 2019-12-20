package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"
)

// Camera struct
type Camera struct {
	camera    *C.Camera
	context   *C.GPContext
	abilities C.CameraAbilities
	config    CameraWidget
	err       error
}

// Init creates a GPhoto2 context, the camera object, inits it, then obtain the camera's abilities and configuration.
// Returns error if any step fails and nil otherwise
func (c *Camera) Init() error {
	c.context = C.gp_context_new()
	p := C.CString("")
	defer C.free(unsafe.Pointer(p))

	C.gp_camera_new(&c.camera)
	if c.err = cameraResultToError(C.gp_camera_init(c.camera, c.context)); c.err != nil {
		return c.err
	} else if c.err = cameraResultToError(C.gp_camera_get_abilities(c.camera, &c.abilities)); c.err != nil {
		return c.err
	} else if c.err = cameraResultToError(C.gp_widget_new(C.GP_WIDGET_WINDOW, p, &c.config.widget)); c.err != nil {
		C.free(unsafe.Pointer(c.config.widget))
		return c.err
	} else if c.err = cameraResultToError(C.gp_camera_get_config(c.camera, &c.config.widget, c.context)); c.err != nil {
		c.config.Free()
		return c.err
	}

	return nil
}

// Exit func
func (c *Camera) Exit() error {
	err := C.gp_camera_exit(c.camera, c.context)
	return cameraResultToError(err)
}

// Cancel func
func (c *Camera) Cancel() {
	C.gp_context_cancel(c.context)
}

// TriggerCapture func
func (c *Camera) TriggerCapture() error {
	err := C.gp_camera_trigger_capture(c.camera, c.context)
	return cameraResultToError(err)
}

// TriggerCaptureToFile func
func (c *Camera) TriggerCaptureToFile() (CameraFilePath, error) {
	var path CameraFilePath
	var _path C.CameraFilePath
	err := cameraResultToError(C.gp_camera_capture(c.camera, captureImage, &_path, c.context))
	if err != nil {
		return path, err
	}
	path.Name = C.GoString(&_path.name[0])
	path.Folder = C.GoString(&_path.folder[0])
	return path, nil
}

// AsyncWaitForEvent func
func (c *Camera) AsyncWaitForEvent(timeout int) chan *CameraEvent {
	var eventType C.CameraEventType
	var vp unsafe.Pointer
	defer C.free(vp)

	ch := make(chan *CameraEvent)

	go func() {
		C.gp_camera_wait_for_event(c.camera, C.int(timeout), &eventType, &vp, c.context)
		ch <- cCameraEventToGoCameraEvent(vp, eventType)
	}()

	return ch
}

// ListFolders func
func (c *Camera) ListFolders(folder string) ([]string, error) {
	if folder == "" {
		folder = "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	cFolder := C.CString(folder)
	defer C.free(unsafe.Pointer(cFolder))

	if err := cameraResultToError(C.gp_camera_folder_list_folders(c.camera, cFolder, cameraList, c.context)); err != nil {
		return []string{}, err
	}
	folderMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(folderMap))
	i := 0
	for key := range folderMap {
		names[i] = key
		i++
	}

	return names, nil
}

// RListFolders func
func (c *Camera) RListFolders(folder string) []string {
	folders := make([]string, 0)
	path := folder
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	subfolders, _ := c.ListFolders(path)
	for _, sub := range subfolders {
		subPath := path + sub
		folders = append(folders, subPath)
		folders = append(folders, c.RListFolders(subPath)...)
	}

	return folders
}

// ListFiles func
func (c *Camera) ListFiles(folder string) ([]string, error) {
	if folder == "" {
		folder = "/"
	}

	if !strings.HasSuffix(folder, "/") {
		folder = folder + "/"
	}

	var cameraList *C.CameraList
	C.gp_list_new(&cameraList)
	defer C.free(unsafe.Pointer(cameraList))

	cFolder := C.CString(folder)
	defer C.free(unsafe.Pointer(cFolder))

	if err := cameraResultToError(C.gp_camera_folder_list_files(c.camera, cFolder, cameraList, c.context)); err != nil {
		return []string{}, err
	}
	fileNameMap, _ := cameraListToMap(cameraList)

	names := make([]string, len(fileNameMap))
	i := 0
	for key := range fileNameMap {
		names[i] = key
		i++
	}

	return names, nil
}

// FileReader func
func (c *Camera) FileReader(folder string, fileName string) io.ReadCloser {
	cfr := new(cameraFileReader)
	cfr.camera = c
	cfr.folder = folder
	cfr.fileName = fileName
	cfr.offset = 0
	cfr.closed = false

	cFileName := C.CString(cfr.fileName)
	cFolderName := C.CString(cfr.folder)
	defer C.free(unsafe.Pointer(cFileName))
	defer C.free(unsafe.Pointer(cFolderName))

	C.gp_file_new(&cfr.cCameraFile)
	C.gp_camera_file_get(c.camera, cFolderName, cFileName, C.GP_FILE_TYPE_NORMAL, cfr.cCameraFile, c.context)

	var cSize C.ulong
	C.gp_file_get_data_and_size(cfr.cCameraFile, &cfr.cBuffer, &cSize)

	cfr.fullSize = uint64(cSize)

	return cfr
}

// DeleteFile func
func (c *Camera) DeleteFile(folder, file string) error {
	folderBytes := []byte(folder)
	fileBytes := []byte(file)
	//Convert the byte arrays into C pointers

	folderPointer := (*C.char)(unsafe.Pointer(&folderBytes[0]))
	filePointer := (*C.char)(unsafe.Pointer(&fileBytes[0]))
	return cameraResultToError(C.gp_camera_file_delete(c.camera, folderPointer, filePointer, c.context))
}

// CapturePreview func
func (c *Camera) CapturePreview() (cf CameraFile, err error) {
	C.gp_file_new(&cf.file)
	if err := cameraResultToError(C.gp_camera_capture_preview(c.camera, cf.file, c.context)); err != nil {
		return cf, err
	}
	if err := cameraResultToError(C.gp_file_get_data_and_size(cf.file, &cf.buf, &cf.cSize)); err != nil {
		return cf, err
	}

	return cf, err
}

// CapturePreviewToFile func
func (c *Camera) CapturePreviewToFile(fileName string) (cf CameraFile, err error) {
	cf, err = c.CapturePreview()
	if err != nil {
		return cf, err
	}

	cfr := new(cameraFileReader)
	cfr.camera = c
	cfr.offset = 0
	cfr.closed = false

	cfr.cCameraFile = cf.file
	cfr.fullSize = uint64(cf.cSize)
	cfr.cBuffer = cf.buf

	defer cfr.Close()

	if fileWriter, err := os.Create(fileName); err == nil {
		io.Copy(fileWriter, cfr)
	}

	return cf, err
}

//
// getters and setters
//

// Model func
func (c *Camera) Model() (string, error) {
	abilities, err := c.Abilities()
	if err != nil {
		return "", err
	}

	return ToString(&abilities.model[0]), nil
}

// ID func
func (c *Camera) ID() (string, error) {
	abilities, err := c.Abilities()
	if err != nil {
		return "", err
	}

	return ToString(&abilities.id[0]), nil
}

// Library func
func (c *Camera) Library() (string, error) {
	abilities, err := c.Abilities()
	if err != nil {
		return "", err
	}

	return ToString(&abilities.library[0]), nil
}

// Abilities func
func (c *Camera) Abilities() (C.CameraAbilities, error) {
	return c.abilities, c.err
}

// SetConfig func
func (c *Camera) SetConfig() error {
	if c.err != nil {
		// something failed during camera init.  Bail!
		return c.err
	}
	if err := cameraResultToError(C.gp_camera_set_config(c.camera, c.config.widget, c.context)); err != nil {
		return fmt.Errorf("error on C.gp_camera_set_config %v", err)
	}
	return nil
}

// Config func
func (c *Camera) Config() (*CameraWidget, error) {
	return &c.config, c.err
}
