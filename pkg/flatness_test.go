package pkg

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestEvaluateFlatness(t *testing.T) {
	data := make([]byte, 84+50)
	binary.LittleEndian.PutUint32(data[80:84], 1)
	put := func(offset int, x, y, z float32) {
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(x))
		binary.LittleEndian.PutUint32(data[offset+4:offset+8], math.Float32bits(y))
		binary.LittleEndian.PutUint32(data[offset+8:offset+12], math.Float32bits(z))
	}
	put(96, 0, 0, 0)
	put(108, 0, 10, 0)
	put(120, 10, 0, 0)

	result, err := evaluateFlatness(data, models.GeometryTest{Type: "is-flat", Side: "bottom", MinContactArea: 1})
	if err != nil || !result.Passed {
		t.Fatalf("expected bottom to pass, got %#v, %v", result, err)
	}
	result, err = evaluateFlatness(data, models.GeometryTest{Type: "is-flat", Side: "top", MinContactArea: 1})
	if err != nil || result.Passed {
		t.Fatalf("expected top to fail, got %#v, %v", result, err)
	}
}

func TestEvaluateFlatnessRejectsUnknownSide(t *testing.T) {
	if _, err := evaluateFlatness(make([]byte, 84), models.GeometryTest{Type: "is-flat", Side: "diagonal"}); err == nil {
		t.Fatal("expected unknown side error")
	}
}

func TestFlatnessLevelUsesProjectedFootprint(t *testing.T) {
	vertices := []flatVertex{{0, 0, 0}, {10, 0, 0}, {0, 10, 0}}
	if got := flatnessLevelMinimum("high", vertices, 2); got != 15 {
		t.Fatalf("high threshold = %v, want 15", got)
	}
	if got := flatnessLevelMinimum("med", vertices, 2); got != 5 {
		t.Fatalf("medium threshold = %v, want 5", got)
	}
}
