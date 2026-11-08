package format

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewASCII(t *testing.T) {
	f := NewASCII()
	if f == nil {
		t.Fatal("NewASCII returned nil")
	}
	if f.float != "%.6f" {
		t.Errorf("Expected float format '%%.6f', got '%s'", f.float)
	}
}

func TestNewASCIIWithVersion(t *testing.T) {
	f := NewASCIIWithVersion(AC1015)
	if f == nil {
		t.Fatal("NewASCIIWithVersion returned nil")
	}
	if f.float != "%.6f" {
		t.Errorf("Expected float format '%%.6f', got '%s'", f.float)
	}
}

func TestASCIIReset(t *testing.T) {
	f := NewASCII()
	f.WriteString(1, "test")
	f.Reset()
	if f.buffer.Len() != 0 {
		t.Errorf("Expected buffer length 0 after reset, got %d", f.buffer.Len())
	}
}

func TestASCIIWriteTo(t *testing.T) {
	f := NewASCII()
	f.WriteString(1, "test")

	var buf bytes.Buffer
	n, err := f.WriteTo(&buf)
	if err != nil {
		t.Errorf("WriteTo failed: %v", err)
	}
	if n <= 0 {
		t.Errorf("Expected positive bytes written, got %d", n)
	}
}

func TestASCIIOutput(t *testing.T) {
	f := NewASCII()
	f.WriteString(1, "hello")
	output := f.Output()
	if output == "" {
		t.Error("Output returned empty string")
	}
	if f.buffer.Len() != 0 {
		t.Errorf("Buffer should be reset after Output(), got length %d", f.buffer.Len())
	}
}

func TestASCIISetPrecision(t *testing.T) {
	f := NewASCII()
	f.SetPrecision(3)
	if f.float != "%.3f" {
		t.Errorf("Expected '%%.3f', got '%s'", f.float)
	}
	f.SetPrecision(8)
	if f.float != "%.8f" {
		t.Errorf("Expected '%%.8f', got '%s'", f.float)
	}
}

func TestASCIIVersion(t *testing.T) {
	f := NewASCII()
	version := f.Version()
	if version != AC1021 {
		t.Errorf("Expected version %s, got %s", AC1021, version)
	}
}

func TestASCIIInt(t *testing.T) {
	f := NewASCII()
	result := f.Int(10, 42)
	expected := "10\n42\n"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestASCIIFloat(t *testing.T) {
	f := NewASCII()
	result := f.Float(10, 3.14159)
	if !strings.Contains(result, "10") {
		t.Errorf("Expected result to contain group code 10, got '%s'", result)
	}
	if !strings.Contains(result, "3.141") {
		t.Errorf("Expected result to contain float value, got '%s'", result)
	}
}

func TestASCIIString(t *testing.T) {
	f := NewASCII()
	result := f.String(1, "test")
	expected := "1\ntest\n"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestASCIIWriteString(t *testing.T) {
	f := NewASCII()
	f.WriteString(1, "hello")
	output := f.Output()
	expected := "1\nhello\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestASCIIWriteInt(t *testing.T) {
	f := NewASCII()
	f.WriteInt(10, 123)
	output := f.Output()
	expected := "10\n123\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestASCIIWriteFloat(t *testing.T) {
	f := NewASCII()
	f.WriteFloat(10, 2.5)
	output := f.Output()
	if !strings.Contains(output, "10") {
		t.Errorf("Expected output to contain '10', got '%s'", output)
	}
	if !strings.Contains(output, "2.5") {
		t.Errorf("Expected output to contain '2.5', got '%s'", output)
	}
}

func TestASCIIWriteBool(t *testing.T) {
	f := NewASCII()
	f.WriteBool(90, true)
	output := f.Output()
	expected := "90\n1\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}

	f.Reset()
	f.WriteBool(90, false)
	output = f.Output()
	expected = "90\n0\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestNewBinary(t *testing.T) {
	f := NewBinary(AC1015)
	if f == nil {
		t.Fatal("NewBinary returned nil")
	}
	if f.dxfversion != AC1015 {
		t.Errorf("Expected version %s, got %s", AC1015, f.dxfversion)
	}
}

func TestBinaryReset(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteString(1, "test")
	f.Reset()
	if f.buffer.Len() != 0 {
		t.Errorf("Expected buffer length 0 after reset, got %d", f.buffer.Len())
	}
}

func TestBinaryWriteTo(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteString(1, "test")

	var buf bytes.Buffer
	n, err := f.WriteTo(&buf)
	if err != nil {
		t.Errorf("WriteTo failed: %v", err)
	}
	if n <= 0 {
		t.Errorf("Expected positive bytes written, got %d", n)
	}
}

func TestBinaryOutput(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteString(1, "test")
	output := f.Output()
	if output == "" {
		t.Error("Output returned empty string")
	}
	if f.buffer.Len() != 0 {
		t.Errorf("Buffer should be reset after Output(), got length %d", f.buffer.Len())
	}
}

func TestBinarySetPrecision(t *testing.T) {
	f := NewBinary(AC1015)
	f.SetPrecision(3)
	if f.float != "%.3f" {
		t.Errorf("Expected '%%.3f', got '%s'", f.float)
	}
}

func TestBinaryVersion(t *testing.T) {
	f := NewBinary(AC1024)
	version := f.Version()
	if version != AC1024 {
		t.Errorf("Expected version %s, got %s", AC1024, version)
	}
}

func TestBinaryString(t *testing.T) {
	f := NewBinary(AC1015)
	result := f.String(1, "test")
	if result != "" {
		t.Error("String() should return empty string for Binary formatter")
	}
}

func TestBinaryInt(t *testing.T) {
	f := NewBinary(AC1015)
	result := f.Int(10, 42)
	if result != "" {
		t.Error("Int() should return empty string for Binary formatter")
	}
}

func TestBinaryFloat(t *testing.T) {
	f := NewBinary(AC1015)
	result := f.Float(10, 3.14)
	if result != "" {
		t.Error("Float() should return empty string for Binary formatter")
	}
}

func TestBinaryWriteString(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteString(1, "test")
	if f.buffer.Len() == 0 {
		t.Error("WriteString should write to buffer")
	}
}

func TestBinaryWriteInt(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteInt(10, 123)
	if f.buffer.Len() == 0 {
		t.Error("WriteInt should write to buffer")
	}
}

func TestBinaryWriteFloat(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteFloat(10, 2.5)
	if f.buffer.Len() == 0 {
		t.Error("WriteFloat should write to buffer")
	}
}

func TestBinaryWriteBool(t *testing.T) {
	f := NewBinary(AC1015)
	f.WriteBool(90, true)
	if f.buffer.Len() == 0 {
		t.Error("WriteBool should write to buffer")
	}
}

func TestBinaryWriteBinaryData(t *testing.T) {
	f := NewBinary(AC1015)
	data := []byte{0x01, 0x02, 0x03, 0x04}
	err := f.WriteBinaryData(310, data)
	if err != nil {
		t.Errorf("WriteBinaryData failed: %v", err)
	}
}

func TestBinaryWriteSignature(t *testing.T) {
	f := NewBinary(AC1015)
	err := f.WriteSignature()
	if err != nil {
		t.Errorf("WriteSignature failed: %v", err)
	}
}

func TestFormatterInterface(t *testing.T) {
	var _ Formatter = NewASCII()
	var _ Formatter = NewBinary(AC1015)
}

func TestDXFVersionConstants(t *testing.T) {
	if AC1009 != "AC1009" {
		t.Error("AC1009 constant mismatch")
	}
	if AC1014 != "AC1014" {
		t.Error("AC1014 constant mismatch")
	}
	if AC1015 != "AC1015" {
		t.Error("AC1015 constant mismatch")
	}
	if AC1018 != "AC1018" {
		t.Error("AC1018 constant mismatch")
	}
	if AC1021 != "AC1021" {
		t.Error("AC1021 constant mismatch")
	}
	if AC1024 != "AC1024" {
		t.Error("AC1024 constant mismatch")
	}
	if AC1027 != "AC1027" {
		t.Error("AC1027 constant mismatch")
	}
	if AC1032 != "AC1032" {
		t.Error("AC1032 constant mismatch")
	}
}

func TestASCIIFloatPrecision(t *testing.T) {
	f := NewASCII()
	f.SetPrecision(2)
	result := f.Float(10, 3.14159)
	if !strings.Contains(result, "3.14") {
		t.Errorf("Expected float with 2 decimal precision, got '%s'", result)
	}
}

func TestASCIIOutputMultiple(t *testing.T) {
	f := NewASCII()
	f.WriteString(0, "SECTION")
	f.WriteInt(2, 100)
	f.WriteString(3, "ENTITIES")
	output := f.Output()
	if !strings.Contains(output, "0\nSECTION") {
		t.Error("Output missing first write")
	}
	if !strings.Contains(output, "2\n100") {
		t.Error("Output missing second write")
	}
	if !strings.Contains(output, "3\nENTITIES") {
		t.Error("Output missing third write")
	}
}

func TestBinaryIntSizes(t *testing.T) {
	f := NewBinary(AC1015)

	f.Int(75, 100)
	if f.buffer.Len() != 4 { // int16 for code 70-79
		t.Errorf("Expected 4 bytes for int16, got %d", f.buffer.Len())
	}

	f.Reset()
	f.Int(95, 100000)
	if f.buffer.Len() != 6 { // int16 + int32 for code 90-99
		t.Errorf("Expected 6 bytes for code 90-99, got %d", f.buffer.Len())
	}
}
