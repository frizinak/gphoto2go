package gphoto2go

// #cgo pkg-config: libgphoto2
// #include <gphoto2.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io"
	"strings"
	"unsafe"
)

// Camera struct
type Camera struct {
	camera  *C.Camera
	context *C.GPContext
}

// Init camera
func (c *Camera) Init() error {
	c.context = C.gp_context_new()

	C.gp_camera_new(&c.camera)
	return cameraResultToError(C.gp_camera_init(c.camera, c.context))
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

// GetAbilities func
func (c *Camera) GetAbilities() (C.CameraAbilities, error) {
	var abilities C.CameraAbilities
	err := cameraResultToError(C.gp_camera_get_abilities(c.camera, &abilities))
	return abilities, err
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
	err := cameraResultToError(C.gp_camera_capture(c.camera, CAPTURE_IMAGE, &_path, c.context))
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
	for key, _ := range fileNameMap {
		names[i] = key
		i += 1
	}

	return names, nil
}

// Model func
func (c *Camera) Model() (string, error) {
	abilities, err := c.GetAbilities()
	if err != nil {
		return "", err
	}
	// model := C.GoString((*C.char)(&abilities.model[0]))
	model := ToString(&abilities.model[0])

	return model, nil
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

func getPreviewFile(file *CameraFile) {
	var cSize C.ulong
	var buf *C.char
	C.gp_file_get_data_and_size(file.file, &buf, &cSize)
}

// CapturePreview func
func (c *Camera) CapturePreview() (cf CameraFile, err error) {
	C.gp_file_new(&cf.file)
	err = cameraResultToError(C.gp_camera_capture_preview(
		c.camera,
		cf.file,
		c.context))
	getPreviewFile(&cf)
	return cf, err

}

// SetConfig func
func (c *Camera) SetConfig(w *CameraWidget) error {
	if err := cameraResultToError(C.gp_camera_set_config(c.camera, w.widget, c.context)); err != nil {
		return fmt.Errorf("error on C.gp_camera_set_config %v", err)
	}
	return nil
}

// GetConfig func
func (c *Camera) GetConfig() (*CameraWidget, error) {
	w := CameraWidget{}

	p := C.CString("")
	defer C.free(unsafe.Pointer(p))

	err := cameraResultToError(C.gp_widget_new(C.GP_WIDGET_WINDOW, p, &w.widget))
	if err != nil {
		C.free(unsafe.Pointer(w.widget))
		return nil, err
	}

	if err := cameraResultToError(C.gp_camera_get_config(c.camera, &w.widget, c.context)); err != nil {
		w.Free()
		return nil, err
	}
	return &w, nil
}