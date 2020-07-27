# gPhoto2-Go

Fork of [heck/gphoto2go](https://github.com/heck/gphoto2go)
which in turn is a fork of [micahwedemeyer/gphoto2go](https://github.com/micahwedemeyer/gphoto2go)

Thanks to them!

# Changes

- Improves errors a bit
- Adds bindings for gp_camera_file_read and gp_camera_file_get_info

I only forked to increase the performance of random reads (for [frizinak/photos](https://github.com/frizinak/photos)).

c.FileReader is pretty slow as gp_camera_file_get reads the entire file in memory,
while c.ReadSeeker allows random access without reading the entire file (gp_camera_file_read) but has the drawback of not knowing the filesize.

# Original README:

A more Go idiomatic interface to the gPhoto2 library.

## Warning

I'm not a great Golang programmer, so this could be a complete disaster. But, if you want to use the gPhoto2 library and you'd prefer to avoid sticking cgo references all over your main Go program, this library might help.

I'm also not a great C programmer, having grown up with garbage collectors and other niceties. Therefore, this library will be riddled with memory leaks. I would appreciate any help in making it more memory efficient.

## Installation

```
go get github.com/micahwedemeyer/gphoto2go
```

## Requirements

You will also need libgphoto2 installed. If you are on Mac OS X, I recommend installing it with homebrew.
```
brew install libgphoto2
```

This should now compile by itself thanks to this `// #cgo pkg-config: libgphoto2`

## Usage

The main goal with this library is to present a Go-friendly interface to the C methods of gPhoto2

### Camera Initializing

```go
camera := new(gphoto2go.Camera)
err := camera.Init()
```

This will create a new Camera struct and intitialize it, which prompts gphoto2 to auto-detect any connected USB cameras.

### Taking a Photo

```go
camera.TriggerCapture()
```
This will trigger the camera.

### Downloading the Photos from the Camera

```go
folders := camera.RListFolders("/")
for _, folder := range folders {
    files, _ := camera.ListFiles(folder)
    for _, fileName := range files {
        cameraFileReader := camera.FileReader(folder, fileName)
        if fileWriter, err := os.Create("/tmp/" + fileName); err == nil {
            io.Copy(fileWriter, cameraFileReader)
        }

        // Important, since there is memory used in the transfer that needs to be freed up
        cameraFileReader.Close()
    }
}
```
### Interpreting errors

Most of the functions will return an error code. If it is less than zero, that means an error has occurred. The library can translate the error integer
into a human readable string.

```go
err := camera.TriggerCapture()
if err < 0 {
    fmt.Printf(gphoto2go.CameraResultToString(err))
}
```
