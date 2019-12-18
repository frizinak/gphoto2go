package gphoto2go

import "testing"

func TestCapturePreviewSanity(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Expected capture to work")
		}
	}()
	cam := &Camera{}
	cam.Init()

	if _, err := cam.CapturePreview(); err != nil {
		t.Fatalf("Expected nil error, got %s. Camera must be on", err)
	}
}
