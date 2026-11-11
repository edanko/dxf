package mixins

import (
	"errors"
	"reflect"
	"testing"

	"github.com/edanko/dxf/entity"
)

func TestSubscriptAttributesBehavior(t *testing.T) {
	mixin := NewSubscriptAttributes()
	mixin.Define("name", "part")
	mixin.Define("count", 3)
	if !mixin.Has("name") || mixin.Has("missing") {
		t.Fatal("Has returned unexpected result")
	}
	value, err := mixin.Get("name")
	if err != nil || value != "part" {
		t.Fatalf("Get = %v, %v", value, err)
	}
	if err := mixin.Set("name", "assembly"); err != nil {
		t.Fatal(err)
	}
	if value, _ := mixin.Get("name"); value != "assembly" {
		t.Fatalf("Set did not update value: %v", value)
	}
	if err := mixin.Set("missing", 1); !errors.Is(err, ErrUnknownAttribute) {
		t.Fatalf("Set unknown error = %v", err)
	}
	if _, err := mixin.Get("missing"); !errors.Is(err, ErrUnknownAttribute) {
		t.Fatalf("Get unknown error = %v", err)
	}
	if !mixin.Delete("count") || mixin.Delete("count") {
		t.Fatal("Delete result is incorrect")
	}
	if got, want := mixin.Keys(), []string{"name"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys = %v, want %v", got, want)
	}
	if mixin.Len() != 1 {
		t.Fatalf("Len = %d, want 1", mixin.Len())
	}
}

func TestSubscriptAttributesCopyAndTypedAccess(t *testing.T) {
	source := map[string]any{"count": 4, "label": "bolt"}
	mixin := NewSubscriptAttributesFrom(source)
	source["count"] = 99
	if value, _ := mixin.Get("count"); value != 4 {
		t.Fatalf("mixin shares input map: %v", value)
	}
	count, err := Attribute[int](mixin, "count")
	if err != nil || count != 4 {
		t.Fatalf("Attribute[int] = %d, %v", count, err)
	}
	if _, err := Attribute[string](mixin, "count"); err == nil {
		t.Fatal("expected type mismatch error")
	}
	copy := mixin.Clone()
	if err := copy.Set("count", 5); err != nil {
		t.Fatal(err)
	}
	if value, _ := mixin.Get("count"); value != 4 {
		t.Fatalf("clone changed original: %v", value)
	}
}

func TestMixinDoesNotChangeEntityData(t *testing.T) {
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 5, 0}
	mixin := NewSubscriptAttributes()
	mixin.Define("label", "edge")
	if err := mixin.Set("label", "changed"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(line.Start, []float64{0, 0, 0}) || !reflect.DeepEqual(line.End, []float64{10, 5, 0}) {
		t.Fatalf("mixin modified entity geometry: %v -> %v", line.Start, line.End)
	}
}

func TestNilMixinIsSafe(t *testing.T) {
	var mixin *SubscriptAttributes
	if mixin.Has("x") || mixin.Len() != 0 || mixin.Keys() != nil {
		t.Fatal("nil mixin accessors should be safe")
	}
	if err := mixin.Set("x", 1); !errors.Is(err, ErrUnknownAttribute) {
		t.Fatalf("nil Set error = %v", err)
	}
	if _, err := mixin.Get("x"); !errors.Is(err, ErrUnknownAttribute) {
		t.Fatalf("nil Get error = %v", err)
	}
}
