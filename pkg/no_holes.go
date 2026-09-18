package pkg

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const noHolesBackgroundTolerance = 8

func isNoHolesTest(test models.GeometryTest) bool {
	return strings.EqualFold(strings.TrimSpace(test.Type), "no-holes")
}

func isDifferentTest(test models.GeometryTest) bool {
	return strings.EqualFold(strings.TrimSpace(test.Type), "is-different")
}

func isMinColorPixelsTest(test models.GeometryTest) bool {
	return strings.EqualFold(strings.TrimSpace(test.Type), "min-color-pixels")
}

func isColorInBoundingBoxTest(test models.GeometryTest) bool {
	return strings.EqualFold(strings.TrimSpace(test.Type), "color-in-bounding-box")
}

func stlGeometryTests(tests []models.GeometryTest) []models.GeometryTest {
	stlTests := make([]models.GeometryTest, 0, len(tests))
	for _, test := range tests {
		if !isNoHolesTest(test) && !isDifferentTest(test) && !isMinColorPixelsTest(test) && !isColorInBoundingBoxTest(test) {
			stlTests = append(stlTests, test)
		}
	}
	return stlTests
}

func boundingBox(test models.GeometryTest, bounds image.Rectangle) (image.Rectangle, error) {
	if len(test.BoundingBox) != 3 {
		return image.Rectangle{}, fmt.Errorf("color-in-bounding-box test requires bounding_box = [left, top, size]")
	}
	left, top, size := test.BoundingBox[0], test.BoundingBox[1], test.BoundingBox[2]
	if left < 0 || top < 0 || size <= 0 || left+size > bounds.Dx() || top+size > bounds.Dy() {
		return image.Rectangle{}, fmt.Errorf("bounding_box [%d, %d, %d] is outside image size %dx%d", left, top, size, bounds.Dx(), bounds.Dy())
	}
	return image.Rect(bounds.Min.X+left, bounds.Min.Y+top, bounds.Min.X+left+size, bounds.Min.Y+top+size), nil
}

func colorMatches(pixel [4]uint8, target [3]uint8) bool {
	const tolerance = 80
	return pixel[3] > 0 && absInt(int(pixel[0])-int(target[0])) <= tolerance && absInt(int(pixel[1])-int(target[1])) <= tolerance && absInt(int(pixel[2])-int(target[2])) <= tolerance
}

func evaluateColorInBoundingBoxImage(path string, test models.GeometryTest) (models.GeometryTestResult, image.Rectangle, error) {
	result := models.GeometryTestResult{Type: test.Type}
	target, err := parseHexColor(test.Color)
	if err != nil {
		return result, image.Rectangle{}, err
	}
	img, err := decodeImage(path)
	if err != nil {
		return result, image.Rectangle{}, fmt.Errorf("decode color-in-bounding-box image: %w", err)
	}
	box, err := boundingBox(test, img.Bounds())
	if err != nil {
		return result, image.Rectangle{}, err
	}
	matching := 0
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			if colorMatches(rgba8(img.At(x, y)), target) {
				matching++
			}
		}
	}
	minimum := test.MinColorPixels
	if minimum <= 0 {
		minimum = 1
	}
	result.Passed = matching >= minimum
	result.Message = fmt.Sprintf("box [%d, %d, %d]: %d pixels near %s (minimum %d)", test.BoundingBox[0], test.BoundingBox[1], test.BoundingBox[2], matching, test.Color, minimum)
	return result, box, nil
}

func drawBoundingBox(canvas *image.RGBA, box image.Rectangle, passed bool) {
	border := color.RGBA{R: 220, G: 53, B: 69, A: 255}
	if passed {
		border = color.RGBA{R: 40, G: 167, B: 69, A: 255}
	}
	for inset := 0; inset < 2; inset++ {
		for x := box.Min.X + inset; x < box.Max.X-inset; x++ {
			canvas.SetRGBA(x, box.Min.Y+inset, border)
			canvas.SetRGBA(x, box.Max.Y-1-inset, border)
		}
		for y := box.Min.Y + inset; y < box.Max.Y-inset; y++ {
			canvas.SetRGBA(box.Min.X+inset, y, border)
			canvas.SetRGBA(box.Max.X-1-inset, y, border)
		}
	}
}

func parseHexColor(value string) ([3]uint8, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 {
		return [3]uint8{}, fmt.Errorf("color must be a #RRGGBB value")
	}
	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return [3]uint8{}, fmt.Errorf("parse color %q: %w", value, err)
	}
	return [3]uint8{uint8(parsed >> 16), uint8(parsed >> 8), uint8(parsed)}, nil
}

func evaluateMinColorPixelsImage(path string, test models.GeometryTest) (models.GeometryTestResult, error) {
	result := models.GeometryTestResult{Type: test.Type}
	target, err := parseHexColor(test.Color)
	if err != nil {
		return result, err
	}
	img, err := decodeImage(path)
	if err != nil {
		return result, fmt.Errorf("decode min-color-pixels image: %w", err)
	}
	const tolerance = 80
	matching := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			pixel := rgba8(img.At(x, y))
			if pixel[3] > 0 && absInt(int(pixel[0])-int(target[0])) <= tolerance && absInt(int(pixel[1])-int(target[1])) <= tolerance && absInt(int(pixel[2])-int(target[2])) <= tolerance {
				matching++
			}
		}
	}
	minimum := test.MinColorPixels
	if minimum <= 0 {
		minimum = 1
	}
	result.Passed = matching >= minimum
	result.Message = fmt.Sprintf("%d pixels near %s (minimum %d)", matching, test.Color, minimum)
	return result, nil
}

func evaluateDifferentImages(firstPath, secondPath string, test models.GeometryTest) (models.GeometryTestResult, error) {
	result := models.GeometryTestResult{Type: test.Type}
	first, err := decodeImage(firstPath)
	if err != nil {
		return result, fmt.Errorf("decode first is-different image: %w", err)
	}
	second, err := decodeImage(secondPath)
	if err != nil {
		return result, fmt.Errorf("decode second is-different image: %w", err)
	}
	if first.Bounds().Size() != second.Bounds().Size() {
		return result, fmt.Errorf("is-different images have different dimensions: %v and %v", first.Bounds().Size(), second.Bounds().Size())
	}

	different := 0
	for y := first.Bounds().Min.Y; y < first.Bounds().Max.Y; y++ {
		for x := first.Bounds().Min.X; x < first.Bounds().Max.X; x++ {
			if pixelsDiffer(first.At(x, y), second.At(x, y)) {
				different++
			}
		}
	}
	minimum := test.MinDifferentPixels
	if minimum <= 0 {
		minimum = 1
	}
	result.Passed = different >= minimum
	result.Message = fmt.Sprintf("%d pixels differ (minimum %d)", different, minimum)
	return result, nil
}

func pixelsDiffer(first, second color.Color) bool {
	firstRGBA, secondRGBA := rgba8(first), rgba8(second)
	for channel := range firstRGBA {
		if absInt(int(firstRGBA[channel])-int(secondRGBA[channel])) > noHolesBackgroundTolerance {
			return true
		}
	}
	return false
}

func decodeImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// evaluateNoHolesImage treats the dominant border colour as the render
// background. Background connected to an image edge is outer space; any other
// background component is an enclosed, see-through hole in a rendered object.
func evaluateNoHolesImage(path string, test models.GeometryTest) (models.GeometryTestResult, error) {
	result := models.GeometryTestResult{Type: test.Type}
	file, err := os.Open(path)
	if err != nil {
		return result, fmt.Errorf("open no-holes image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return result, fmt.Errorf("decode no-holes image: %w", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() < 3 || bounds.Dy() < 3 {
		return result, fmt.Errorf("no-holes image must be at least 3 by 3 pixels")
	}

	background := dominantBorderColor(img)
	width, height := bounds.Dx(), bounds.Dy()
	isBackground := make([]bool, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			isBackground[y*width+x] = nearBackground(img.At(bounds.Min.X+x, bounds.Min.Y+y), background)
		}
	}

	exterior := make([]bool, len(isBackground))
	queue := make([]int, 0, 2*(width+height))
	pushExterior := func(x, y int) {
		index := y*width + x
		if isBackground[index] && !exterior[index] {
			exterior[index] = true
			queue = append(queue, index)
		}
	}
	for x := 0; x < width; x++ {
		pushExterior(x, 0)
		pushExterior(x, height-1)
	}
	for y := 1; y < height-1; y++ {
		pushExterior(0, y)
		pushExterior(width-1, y)
	}
	for head := 0; head < len(queue); head++ {
		index := queue[head]
		x, y := index%width, index/width
		for _, offset := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nextX, nextY := x+offset[0], y+offset[1]
			if nextX >= 0 && nextX < width && nextY >= 0 && nextY < height {
				pushExterior(nextX, nextY)
			}
		}
	}

	minimum := test.MinHolePixels
	if minimum <= 0 {
		minimum = 1
	}
	visited := make([]bool, len(isBackground))
	largest := 0
	for start, backgroundPixel := range isBackground {
		if !backgroundPixel || exterior[start] || visited[start] {
			continue
		}
		visited[start] = true
		queue = append(queue[:0], start)
		for head := 0; head < len(queue); head++ {
			index := queue[head]
			x, y := index%width, index/width
			for _, offset := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nextX, nextY := x+offset[0], y+offset[1]
				if nextX < 0 || nextX >= width || nextY < 0 || nextY >= height {
					continue
				}
				next := nextY*width + nextX
				if isBackground[next] && !exterior[next] && !visited[next] {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
		if len(queue) > largest {
			largest = len(queue)
		}
	}

	result.Passed = largest < minimum
	if result.Passed {
		result.Message = fmt.Sprintf("no enclosed background regions of %d or more pixels", minimum)
	} else {
		result.Message = fmt.Sprintf("found enclosed background region of %d pixels (limit %d)", largest, minimum-1)
	}
	return result, nil
}

func runImageTests(images []models.GenerateImageResult, tests []models.GeometryTest) ([]models.GeometryTestResult, error) {
	results := make([]models.GeometryTestResult, 0, len(tests))
	for _, test := range tests {
		if !isNoHolesTest(test) {
			continue
		}
		camera := strings.TrimSpace(test.Camera)
		if camera == "" {
			return results, fmt.Errorf("no-holes test requires camera")
		}
		var path string
		for _, imageResult := range images {
			if imageResult.CameraName == camera {
				path = imageResult.OutputPath
				break
			}
		}
		if path == "" {
			return results, fmt.Errorf("no-holes test image for camera %q was not generated", camera)
		}
		result, err := evaluateNoHolesImage(path, test)
		if err != nil {
			return results, err
		}
		result.Message = fmt.Sprintf("%s camera: %s", camera, result.Message)
		results = append(results, result)
	}
	return results, nil
}

func runIsDifferentTests(images []models.GenerateImageResult, tests []models.GeometryTest) ([]models.GeometryTestResult, error) {
	results := make([]models.GeometryTestResult, 0, len(tests))
	for testIndex, test := range tests {
		if !isDifferentTest(test) {
			continue
		}
		paths := map[string]string{}
		for _, image := range images {
			paths[image.CameraName] = image.OutputPath
		}
		first := paths[isDifferentCameraName(testIndex, "base")]
		second := paths[isDifferentCameraName(testIndex, "different")]
		if first == "" || second == "" {
			return results, fmt.Errorf("is-different test for instance_target %q did not generate both images", test.InstanceTarget)
		}
		result, err := evaluateDifferentImages(first, second, test)
		if err != nil {
			return results, err
		}
		result.Message = fmt.Sprintf("%s: %s", test.InstanceTarget, result.Message)
		results = append(results, result)
	}
	return results, nil
}

func runMinColorPixelsTests(images []models.GenerateImageResult, tests []models.GeometryTest) ([]models.GeometryTestResult, error) {
	results := make([]models.GeometryTestResult, 0, len(tests))
	for _, test := range tests {
		if !isMinColorPixelsTest(test) {
			continue
		}
		if strings.TrimSpace(test.InstanceTarget) == "" || strings.TrimSpace(test.Camera) == "" {
			return results, fmt.Errorf("min-color-pixels test requires instance_target and camera")
		}
		var path string
		for _, image := range images {
			if image.InstanceConfig.Name == test.InstanceTarget && image.CameraName == test.Camera {
				path = image.OutputPath
				break
			}
		}
		if path == "" {
			return results, fmt.Errorf("min-color-pixels test image for instance_target %q and camera %q was not generated", test.InstanceTarget, test.Camera)
		}
		result, err := evaluateMinColorPixelsImage(path, test)
		if err != nil {
			return results, err
		}
		result.Message = fmt.Sprintf("%s/%s: %s", test.InstanceTarget, test.Camera, result.Message)
		results = append(results, result)
	}
	return results, nil
}

// runColorInBoundingBoxTests checks requested square regions and writes one
// diagnostic copy per source render. Tests for the same render are composed on
// that copy, so a report can show several green/red expectations together.
func runColorInBoundingBoxTests(images []models.GenerateImageResult, tests []models.GeometryTest) ([]models.GeometryTestResult, error) {
	results := make([]models.GeometryTestResult, 0, len(tests))
	byImage := make(map[int][]models.GeometryTest)
	for _, test := range tests {
		if !isColorInBoundingBoxTest(test) {
			continue
		}
		if strings.TrimSpace(test.InstanceTarget) == "" || strings.TrimSpace(test.Camera) == "" {
			return results, fmt.Errorf("color-in-bounding-box test requires instance_target and camera")
		}
		index := -1
		for imageIndex, imageResult := range images {
			if imageResult.InstanceConfig.Name == test.InstanceTarget && imageResult.CameraName == test.Camera {
				index = imageIndex
				break
			}
		}
		if index < 0 {
			return results, fmt.Errorf("color-in-bounding-box test image for instance_target %q and camera %q was not generated", test.InstanceTarget, test.Camera)
		}
		byImage[index] = append(byImage[index], test)
	}

	for imageIndex, imageTests := range byImage {
		path := images[imageIndex].OutputPath
		img, err := decodeImage(path)
		if err != nil {
			return results, fmt.Errorf("decode color-in-bounding-box diagnostic image: %w", err)
		}
		canvas := image.NewRGBA(img.Bounds())
		draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src)
		diagnosticPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".test-results.png"
		for _, test := range imageTests {
			result, box, err := evaluateColorInBoundingBoxImage(path, test)
			if err != nil {
				return results, err
			}
			drawBoundingBox(canvas, box, result.Passed)
			result.DiagnosticImagePath = diagnosticPath
			images[imageIndex].Tests = append(images[imageIndex].Tests, result)
			results = append(results, result)
		}
		if err := writePNG(diagnosticPath, canvas); err != nil {
			return results, err
		}
	}
	return results, nil
}

func writePNG(path string, img image.Image) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create diagnostic image directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create diagnostic image: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("encode diagnostic image: %w", err)
	}
	return nil
}

func dominantBorderColor(img image.Image) [4]uint8 {
	bounds := img.Bounds()
	counts := map[[4]uint8]int{}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		counts[rgba8(img.At(x, bounds.Min.Y))]++
		counts[rgba8(img.At(x, bounds.Max.Y-1))]++
	}
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		counts[rgba8(img.At(bounds.Min.X, y))]++
		counts[rgba8(img.At(bounds.Max.X-1, y))]++
	}
	var background [4]uint8
	most := -1
	for colour, count := range counts {
		if count > most {
			background, most = colour, count
		}
	}
	return background
}

func rgba8(colour color.Color) [4]uint8 {
	r, g, b, a := colour.RGBA()
	return [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func nearBackground(colour color.Color, background [4]uint8) bool {
	pixel := rgba8(colour)
	if pixel[3] == 0 {
		return true
	}
	for channel := range background {
		if absInt(int(pixel[channel])-int(background[channel])) > noHolesBackgroundTolerance {
			return false
		}
	}
	return true
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
