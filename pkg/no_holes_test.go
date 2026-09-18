package pkg

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestEvaluateNoHolesImage(t *testing.T) {
	tests := []struct {
		name   string
		draw   func(*image.RGBA)
		passed bool
	}{
		{
			name: "solid disconnected objects keep the outer background valid",
			draw: func(img *image.RGBA) {
				draw.Draw(img, image.Rect(2, 2, 7, 8), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(10, 3, 15, 9), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
			},
			passed: true,
		},
		{
			name: "enclosed background is a see-through hole",
			draw: func(img *image.RGBA) {
				draw.Draw(img, image.Rect(2, 2, 14, 10), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
				draw.Draw(img, image.Rect(6, 4, 10, 8), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
			},
			passed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeNoHolesTestImage(t, tt.draw)
			result, err := evaluateNoHolesImage(path, models.GeometryTest{Type: "no-holes"})
			if err != nil {
				t.Fatal(err)
			}
			if result.Passed != tt.passed {
				t.Fatalf("passed = %t, want %t (%s)", result.Passed, tt.passed, result.Message)
			}
		})
	}
}

func TestRunImageTestsRequiresRequestedCamera(t *testing.T) {
	_, err := runImageTests(nil, []models.GeometryTest{{Type: "no-holes", Camera: "top"}})
	if err == nil {
		t.Fatal("expected missing camera image error")
	}
}

func TestEvaluateDifferentImages(t *testing.T) {
	base := writeNoHolesTestImage(t, func(img *image.RGBA) {})
	different := writeDifferentTestImage(t, func(img *image.RGBA) {
		draw.Draw(img, image.Rect(2, 2, 6, 6), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)
	})
	result, err := evaluateDifferentImages(base, different, models.GeometryTest{Type: "is-different", MinDifferentPixels: 16})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed {
		t.Fatalf("expected images to differ: %s", result.Message)
	}

	result, err = evaluateDifferentImages(base, different, models.GeometryTest{Type: "is-different", MinDifferentPixels: 17})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed {
		t.Fatalf("expected threshold failure: %s", result.Message)
	}
}

func TestEvaluateMinColorPixelsImage(t *testing.T) {
	path := writeTestImage(t, "red.png", func(img *image.RGBA) {
		draw.Draw(img, image.Rect(2, 2, 6, 6), &image.Uniform{C: color.RGBA{R: 240, G: 30, B: 30, A: 255}}, image.Point{}, draw.Src)
	})
	result, err := evaluateMinColorPixelsImage(path, models.GeometryTest{Type: "min-color-pixels", Color: "#ff0000", MinColorPixels: 16})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed {
		t.Fatalf("expected red coverage to pass: %s", result.Message)
	}
	result, err = evaluateMinColorPixelsImage(path, models.GeometryTest{Type: "min-color-pixels", Color: "#ff0000", MinColorPixels: 17})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed {
		t.Fatalf("expected red coverage threshold failure: %s", result.Message)
	}
}

func TestRunColorInBoundingBoxTestsDrawsCombinedDiagnostic(t *testing.T) {
	path := writeTestImage(t, "render.png", func(img *image.RGBA) {
		draw.Draw(img, image.Rect(2, 2, 6, 6), &image.Uniform{C: color.RGBA{R: 240, G: 30, B: 30, A: 255}}, image.Point{}, draw.Src)
	})
	images := []models.GenerateImageResult{{
		InstanceConfig: models.InstanceConfig{Name: "fixture"},
		CameraName:     "top",
		OutputPath:     path,
	}}
	results, err := runColorInBoundingBoxTests(images, []models.GeometryTest{
		{Type: "color-in-bounding-box", InstanceTarget: "fixture", Camera: "top", Color: "#ff0000", BoundingBox: []int{2, 2, 4}},
		{Type: "color-in-bounding-box", InstanceTarget: "fixture", Camera: "top", Color: "#ff0000", BoundingBox: []int{10, 2, 4}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || !results[0].Passed || results[1].Passed {
		t.Fatalf("unexpected test results: %#v", results)
	}
	if results[0].DiagnosticImagePath == "" || results[0].DiagnosticImagePath != results[1].DiagnosticImagePath {
		t.Fatalf("expected both regions in one diagnostic image: %#v", results)
	}
	diagnostic, err := decodeImage(results[0].DiagnosticImagePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := rgba8(diagnostic.At(2, 2)); got != [4]uint8{40, 167, 69, 255} {
		t.Fatalf("passing border = %v, want green", got)
	}
	if got := rgba8(diagnostic.At(10, 2)); got != [4]uint8{220, 53, 69, 255} {
		t.Fatalf("failing border = %v, want red", got)
	}
}

func writeNoHolesTestImage(t *testing.T, paint func(*image.RGBA)) string {
	return writeTestImage(t, "render.png", paint)
}

func writeDifferentTestImage(t *testing.T, paint func(*image.RGBA)) string {
	return writeTestImage(t, "different.png", paint)
}

func writeTestImage(t *testing.T, name string, paint func(*image.RGBA)) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 18, 12))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	paint(img)
	path := filepath.Join(t.TempDir(), name)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
	return path
}
