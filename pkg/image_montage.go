package pkg

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const allInstancesImageType = "all_instances"

func generateAllInstancesMontages(config *models.Config, images []models.GenerateImageResult) ([]models.GenerateImageResult, error) {
	var results []models.GenerateImageResult
	for _, imageConfig := range config.Design.ExportImages {
		if imageConfig.ImageType != allInstancesImageType {
			continue
		}

		paths := make([]string, 0)
		for _, image := range images {
			if image.CameraName == imageConfig.CameraName && image.OutputPath != "" {
				paths = append(paths, image.OutputPath)
			}
		}
		if len(paths) == 0 {
			continue
		}

		outputPath := filepath.Join(filepath.Dir(config.ConfigFile), "img", config.Design.Version, "all-instances-"+imageConfig.CameraName+".png")
		if err := createImageMontage(paths, outputPath, imageConfig.ImageSize); err != nil {
			return results, fmt.Errorf("create all-instances %q montage: %w", imageConfig.CameraName, err)
		}

		results = append(results, models.GenerateImageResult{
			OutputPath:   outputPath,
			Command:      "all_instances montage",
			CameraName:   "all-instances-" + imageConfig.CameraName,
			CameraCoords: imageConfig.CameraCoordinates,
		})
	}
	return results, nil
}

func createImageMontage(paths []string, outputPath, imageSize string) error {
	width, height, err := parseMontageImageSize(imageSize)
	if err != nil {
		return err
	}
	columns := int(math.Ceil(math.Sqrt(float64(len(paths)))))
	rows := int(math.Ceil(float64(len(paths)) / float64(columns)))
	cellWidth, cellHeight := width/columns, height/rows
	if cellWidth == 0 || cellHeight == 0 {
		return fmt.Errorf("image size %dx%d is too small for %d instances", width, height, len(paths))
	}

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			canvas.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	for i, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open source image %q: %w", path, err)
		}
		source, _, decodeErr := image.Decode(file)
		closeErr := file.Close()
		if decodeErr != nil {
			return fmt.Errorf("decode source image %q: %w", path, decodeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close source image %q: %w", path, closeErr)
		}

		sourceBounds := source.Bounds()
		scale := math.Min(float64(cellWidth)/float64(sourceBounds.Dx()), float64(cellHeight)/float64(sourceBounds.Dy()))
		drawWidth := int(math.Round(float64(sourceBounds.Dx()) * scale))
		drawHeight := int(math.Round(float64(sourceBounds.Dy()) * scale))
		x := (i%columns)*cellWidth + (cellWidth-drawWidth)/2
		y := (i/columns)*cellHeight + (cellHeight-drawHeight)/2
		scaleImageNearest(canvas, image.Rect(x, y, x+drawWidth, y+drawHeight), source, sourceBounds)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create montage directory for %q: %w", outputPath, err)
	}
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create montage %q: %w", outputPath, err)
	}
	encodeErr := png.Encode(output, canvas)
	closeErr := output.Close()
	if encodeErr != nil {
		return fmt.Errorf("encode montage %q: %w", outputPath, encodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close montage %q: %w", outputPath, closeErr)
	}
	return nil
}

func scaleImageNearest(destination *image.RGBA, destinationBounds image.Rectangle, source image.Image, sourceBounds image.Rectangle) {
	for y := destinationBounds.Min.Y; y < destinationBounds.Max.Y; y++ {
		sourceY := sourceBounds.Min.Y + (y-destinationBounds.Min.Y)*sourceBounds.Dy()/destinationBounds.Dy()
		for x := destinationBounds.Min.X; x < destinationBounds.Max.X; x++ {
			sourceX := sourceBounds.Min.X + (x-destinationBounds.Min.X)*sourceBounds.Dx()/destinationBounds.Dx()
			sourceColor := color.NRGBAModel.Convert(source.At(sourceX, sourceY)).(color.NRGBA)
			if sourceColor.A == 0 {
				continue
			}
			if sourceColor.A == 255 {
				destination.SetRGBA(x, y, color.RGBA{R: sourceColor.R, G: sourceColor.G, B: sourceColor.B, A: 255})
				continue
			}
			destinationColor := destination.RGBAAt(x, y)
			alpha := uint32(sourceColor.A)
			destination.SetRGBA(x, y, color.RGBA{
				R: uint8((uint32(sourceColor.R)*alpha + uint32(destinationColor.R)*(255-alpha)) / 255),
				G: uint8((uint32(sourceColor.G)*alpha + uint32(destinationColor.G)*(255-alpha)) / 255),
				B: uint8((uint32(sourceColor.B)*alpha + uint32(destinationColor.B)*(255-alpha)) / 255),
				A: 255,
			})
		}
	}
}

func parseMontageImageSize(imageSize string) (int, int, error) {
	if imageSize == "" {
		return 1920, 1080, nil
	}
	parts := strings.FieldsFunc(strings.ToLower(imageSize), func(r rune) bool { return r == ',' || r == 'x' })
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid image_size %q; expected WIDTH,HEIGHT or WIDTHxHEIGHT", imageSize)
	}
	width, widthErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, heightErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid image_size %q; dimensions must be positive integers", imageSize)
	}
	return width, height, nil
}
