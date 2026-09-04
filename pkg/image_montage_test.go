package pkg

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestGenerateAllInstancesMontages(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first-nice.png")
	second := filepath.Join(dir, "second-nice.png")
	writeSolidPNG(t, first, color.RGBA{R: 255, A: 255})
	writeSolidPNG(t, second, color.RGBA{B: 255, A: 255})

	config := &models.Config{
		ConfigFile: filepath.Join(dir, "config.toml"),
		Design: models.DesignConfig{
			Version: "v0.1",
			ExportImages: []models.ExportCameraCoordinates{{
				CameraName: "nice",
				ImageType:  allInstancesImageType,
				ImageSize:  "20x10",
			}},
		},
	}
	results, err := generateAllInstancesMontages(config, []models.GenerateImageResult{
		{CameraName: "nice", OutputPath: first},
		{CameraName: "nice", OutputPath: second},
		{CameraName: "front", OutputPath: first},
	})
	if err != nil {
		t.Fatalf("generateAllInstancesMontages: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one montage, got %d", len(results))
	}
	if results[0].CameraName != "all-instances-nice" {
		t.Fatalf("unexpected montage camera name: %q", results[0].CameraName)
	}

	output, err := os.Open(results[0].OutputPath)
	if err != nil {
		t.Fatalf("open montage: %v", err)
	}
	defer output.Close()
	montage, err := png.Decode(output)
	if err != nil {
		t.Fatalf("decode montage: %v", err)
	}
	if got := montage.Bounds().Size(); got != (image.Point{X: 20, Y: 10}) {
		t.Fatalf("unexpected montage size: %v", got)
	}
	if got := color.RGBAModel.Convert(montage.At(2, 5)).(color.RGBA); got.R != 255 || got.B != 0 {
		t.Fatalf("left cell does not contain the first image: %#v", got)
	}
	if got := color.RGBAModel.Convert(montage.At(17, 5)).(color.RGBA); got.B != 255 || got.R != 0 {
		t.Fatalf("right cell does not contain the second image: %#v", got)
	}
}

func writeSolidPNG(t *testing.T, path string, fill color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create source png: %v", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode source png: %v", err)
	}
}
