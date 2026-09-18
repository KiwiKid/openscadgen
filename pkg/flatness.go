package pkg

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

type flatVertex struct{ x, y, z float64 }

func flatAxis(side string) (int, float64, error) {
	switch strings.ToLower(strings.TrimSpace(side)) {
	case "bottom":
		return 2, -1, nil
	case "top":
		return 2, 1, nil
	case "front":
		return 1, -1, nil
	case "back":
		return 1, 1, nil
	case "left":
		return 0, -1, nil
	case "right":
		return 0, 1, nil
	default:
		return 0, 0, fmt.Errorf("is-flat side must be top, bottom, front, back, left, or right")
	}
}

func evaluateFlatness(data []byte, test models.GeometryTest) (models.GeometryTestResult, error) {
	result := models.GeometryTestResult{Type: test.Type}
	if strings.ToLower(strings.TrimSpace(test.Type)) != "is-flat" {
		return result, fmt.Errorf("unsupported geometry test %q", test.Type)
	}
	axis, direction, err := flatAxis(test.Side)
	if err != nil {
		return result, err
	}
	vertices, areas, err := parseBinarySTL(data)
	if err != nil {
		vertices, areas, err = parseASCIISTL(data)
	}
	if err != nil {
		return result, err
	}
	if len(vertices) == 0 {
		return result, fmt.Errorf("STL is too small")
	}
	value := func(v flatVertex) float64 {
		if axis == 0 {
			return v.x
		}
		if axis == 1 {
			return v.y
		}
		return v.z
	}
	extreme := value(vertices[0])
	for _, v := range vertices[1:] {
		if direction*value(v) > direction*extreme {
			extreme = value(v)
		}
	}
	tolerance := test.Tolerance
	if tolerance <= 0 {
		tolerance = 0.01
	}
	total, contact := 0.0, 0.0
	for i, area := range areas {
		total += area
		near := true
		for _, v := range vertices[i*3 : i*3+3] {
			if math.Abs(value(v)-extreme) > tolerance {
				near = false
				break
			}
		}
		if near && outwardNormal(vertices[i*3], vertices[i*3+1], vertices[i*3+2], axis, direction) {
			contact += area
		}
	}
	minimum := test.MinContactArea
	if minimum <= 0 {
		minimum = flatnessLevelMinimum(test.Level, vertices, axis)
		if minimum < 0 {
			return result, fmt.Errorf("is-flat level must be low, medium, med, or high")
		}
		if minimum <= 0 {
			minimum = total * 0.05
		}
	}
	result.Passed = contact >= minimum
	result.Message = fmt.Sprintf("%s contact area %.3f mm² (required %.3f)", test.Side, contact, minimum)
	return result, nil
}

func flatnessLevelMinimum(level string, vertices []flatVertex, axis int) float64 {
	level = strings.ToLower(strings.TrimSpace(level))
	if level == "" {
		return 0
	}
	minimumRatio := map[string]float64{"low": 0.01, "medium": 0.05, "med": 0.05, "high": 0.15}[level]
	if minimumRatio == 0 {
		return -1
	}
	var minA, maxA, minB, maxB float64
	first := true
	for _, v := range vertices {
		a, b := v.x, v.y
		if axis == 0 {
			a, b = v.y, v.z
		} else if axis == 1 {
			a, b = v.x, v.z
		}
		if first {
			minA, maxA, minB, maxB = a, a, b, b
			first = false
		} else {
			minA = math.Min(minA, a)
			maxA = math.Max(maxA, a)
			minB = math.Min(minB, b)
			maxB = math.Max(maxB, b)
		}
	}
	return (maxA - minA) * (maxB - minB) * minimumRatio
}

func parseBinarySTL(data []byte) ([]flatVertex, []float64, error) {
	if len(data) < 84 {
		return nil, nil, fmt.Errorf("not binary STL")
	}
	count := int(binary.LittleEndian.Uint32(data[80:84]))
	if count == 0 || len(data) < 84+count*50 {
		return nil, nil, fmt.Errorf("not binary STL")
	}
	vertices := make([]flatVertex, count*3)
	areas := make([]float64, count)
	for i := 0; i < count; i++ {
		off := 84 + i*50
		for j := 0; j < 3; j++ {
			p := off + 12 + j*12
			vertices[i*3+j] = flatVertex{float64(math.Float32frombits(binary.LittleEndian.Uint32(data[p : p+4]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(data[p+4 : p+8]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(data[p+8 : p+12])))}
		}
		areas[i] = triangleArea(vertices[i*3], vertices[i*3+1], vertices[i*3+2])
	}
	return vertices, areas, nil
}

func parseASCIISTL(data []byte) ([]flatVertex, []float64, error) {
	var vertices []flatVertex
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 4 && fields[0] == "vertex" {
			x, e1 := strconv.ParseFloat(fields[1], 64)
			y, e2 := strconv.ParseFloat(fields[2], 64)
			z, e3 := strconv.ParseFloat(fields[3], 64)
			if e1 != nil || e2 != nil || e3 != nil {
				return nil, nil, fmt.Errorf("invalid ASCII STL vertex")
			}
			vertices = append(vertices, flatVertex{x, y, z})
		}
	}
	if scanner.Err() != nil || len(vertices) == 0 || len(vertices)%3 != 0 {
		return nil, nil, fmt.Errorf("invalid or empty ASCII STL")
	}
	areas := make([]float64, len(vertices)/3)
	for i := range areas {
		areas[i] = triangleArea(vertices[i*3], vertices[i*3+1], vertices[i*3+2])
	}
	return vertices, areas, nil
}

func triangleArea(a, b, c flatVertex) float64 {
	x1, y1, z1 := b.x-a.x, b.y-a.y, b.z-a.z
	x2, y2, z2 := c.x-a.x, c.y-a.y, c.z-a.z
	return 0.5 * math.Sqrt((y1*z2-z1*y2)*(y1*z2-z1*y2)+(z1*x2-x1*z2)*(z1*x2-x1*z2)+(x1*y2-y1*x2)*(x1*y2-y1*x2))
}

func outwardNormal(a, b, c flatVertex, axis int, direction float64) bool {
	x1, y1, z1 := b.x-a.x, b.y-a.y, b.z-a.z
	x2, y2, z2 := c.x-a.x, c.y-a.y, c.z-a.z
	n := []float64{y1*z2 - z1*y2, z1*x2 - x1*z2, x1*y2 - y1*x2}
	return direction*n[axis] > 0
}

func runGeometryTests(path string, tests []models.GeometryTest) ([]models.GeometryTestResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	results := make([]models.GeometryTestResult, 0, len(tests))
	for _, test := range tests {
		result, err := evaluateFlatness(data, test)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}
