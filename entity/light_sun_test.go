package entity

import (
	"testing"
)

func TestLightEntity(t *testing.T) {
	t.Run("NewLight", func(t *testing.T) {
		light := NewLight()
		if light.Type != LightTypeDistant {
			t.Errorf("Expected default light type to be Distant, got %v", light.Type)
		}
		if !light.IsOn() {
			t.Error("Light should be on by default")
		}
		if light.Intensity != 1.0 {
			t.Errorf("Expected default intensity 1.0, got %f", light.Intensity)
		}
	})

	t.Run("NewPointLight", func(t *testing.T) {
		light := NewPointLight("TestLight", [3]float64{1, 2, 3}, 0.8)
		if light.Type != LightTypePoint {
			t.Errorf("Expected point light type, got %v", light.Type)
		}
		if light.Location[0] != 1 || light.Location[1] != 2 || light.Location[2] != 3 {
			t.Error("Point light location not set correctly")
		}
		if light.Intensity != 0.8 {
			t.Errorf("Expected intensity 0.8, got %f", light.Intensity)
		}
	})

	t.Run("NewSpotLight", func(t *testing.T) {
		light := NewSpotLight("TestLight", [3]float64{0, 0, 0}, [3]float64{1, 1, 1}, 30.0, 45.0)
		if light.Type != LightTypeSpot {
			t.Errorf("Expected spot light type, got %v", light.Type)
		}
		if light.HotspotAngle != 30.0 {
			t.Errorf("Expected hotspot angle 30.0, got %f", light.HotspotAngle)
		}
	})

	t.Run("LightValidation", func(t *testing.T) {
		light := NewLight()
		light.Name = "TestLight" // Set required name

		light.Intensity = -0.5 // Invalid
		errors := light.Validate()
		if len(errors) == 0 {
			t.Error("Should have validation errors for negative intensity")
		}

		light.Intensity = 0.5 // Valid
		errors = light.Validate()
		if len(errors) != 0 {
			t.Errorf("Should have no validation errors for valid intensity: %v", errors)
		}
	})

	t.Run("LightMethods", func(t *testing.T) {
		light := NewLight()
		light.TurnOff()
		if light.IsOn() {
			t.Error("Light should be off")
		}

		light.TurnOn()
		if !light.IsOn() {
			t.Error("Light should be on")
		}

		light.SetLocation(10, 20, 30)
		if light.Location[0] != 10 || light.Location[1] != 20 || light.Location[2] != 30 {
			t.Error("Location not set correctly")
		}
	})
}

func TestSunEntity(t *testing.T) {
	t.Run("NewSun", func(t *testing.T) {
		sun := NewSun()
		if !sun.IsOn() {
			t.Error("Sun should be on by default")
		}
		if sun.Intensity != 1.0 {
			t.Errorf("Expected default intensity 1.0, got %f", sun.Intensity)
		}
		if sun.JulianDay != 2456922 {
			t.Errorf("Expected default Julian day 2456922, got %d", sun.JulianDay)
		}
	})

	t.Run("NewSunWithDateTime", func(t *testing.T) {
		sun := NewSunWithDateTime(2023, 6, 15, 12, 30, 0)
		if sun.JulianDay < 2450000 || sun.JulianDay > 2500000 {
			t.Errorf("Unexpected Julian day for date: %d", sun.JulianDay)
		}
		if sun.Time != 45000.0 {
			t.Errorf("Expected time 45000.0, got %f", sun.Time)
		}
	})

	t.Run("SunValidation", func(t *testing.T) {
		sun := NewSun()
		sun.Intensity = 1.5 // Invalid
		errors := sun.Validate()
		if len(errors) == 0 {
			t.Error("Should have validation errors for invalid intensity")
		}

		sun.Intensity = 0.5 // Valid
		errors = sun.Validate()
		if len(errors) != 0 {
			t.Errorf("Should have no validation errors for valid intensity: %v", errors)
		}
	})

	t.Run("SunMethods", func(t *testing.T) {
		sun := NewSun()
		sun.TurnOff()
		if sun.IsOn() {
			t.Error("Sun should be off")
		}

		sun.TurnOn()
		if !sun.IsOn() {
			t.Error("Sun should be on")
		}

		sun.SetTime(14, 30, 45)
		if sun.Time != 52245.0 {
			t.Errorf("Expected time 52245.0, got %f", sun.Time)
		}

		hours, minutes, _ := sun.GetTimeHoursMinutes()
		if hours != 14 || minutes != 30 {
			t.Errorf("Expected time 14:30, got %02d:%02d", hours, minutes)
		}
	})
}

func TestLightTypes(t *testing.T) {
	t.Run("LightTypeString", func(t *testing.T) {
		tests := map[LightType]string{
			LightTypeDistant: "Distant",
			LightTypePoint:   "Point",
			LightTypeSpot:    "Spot",
		}

		for lt, expected := range tests {
			if lt.String() != expected {
				t.Errorf("Expected %s, got %s", expected, lt.String())
			}
		}
	})
}
