package acisbrowser

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestExtractEntityRegionSAT(t *testing.T) {
	region := entity.NewRegion()
	region.SetHandleString("A1")
	region.SetSATLines([]string{"700 0 1 0", "body $-1 $1 $-1 $-1 #"})
	region.SetACISInfo(entity.ACISMetadata{
		Format:      entity.ACISFormatSAT,
		Version:     700,
		RecordCount: 2,
	})

	got, ok := ExtractEntity(region)
	if !ok {
		t.Fatal("expected SAT region to be extracted")
	}
	if got.Type != "REGION" {
		t.Fatalf("type = %q, want REGION", got.Type)
	}
	if got.Format != "SAT" {
		t.Fatalf("format = %q, want SAT", got.Format)
	}
	if len(got.Lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(got.Lines))
	}
	if got.Handle == "" {
		t.Fatal("expected a handle")
	}
}

func TestExtractEntityBodyBinary(t *testing.T) {
	body := entity.NewBody()
	payload := []byte("SAB payload")
	body.SATData = payload
	body.SetACISInfo(entity.ACISMetadata{
		Format:    entity.ACISFormatSAB,
		BodyCount: 1,
	})

	got, ok := ExtractEntity(body)
	if !ok {
		t.Fatal("expected binary body to be extracted")
	}
	if got.Type != "BODY" {
		t.Fatalf("type = %q, want BODY", got.Type)
	}
	if got.Format != "SAB" {
		t.Fatalf("format = %q, want SAB", got.Format)
	}
	if string(got.Binary) != string(payload) {
		t.Fatalf("binary payload = %q, want %q", got.Binary, payload)
	}
}

func TestExtractEntitySkipsNonACIS(t *testing.T) {
	if _, ok := ExtractEntity(entity.NewLine()); ok {
		t.Fatal("line should not be extracted as ACIS data")
	}
}

func TestExtractFromDrawing(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	region := entity.NewRegion()
	region.SetSATLines([]string{"700 0 1 0"})
	region.SetACISInfo(entity.ACISMetadata{Format: entity.ACISFormatSAT, Version: 700})
	if err := doc.AddEntity(region); err != nil {
		t.Fatalf("add region: %v", err)
	}

	got := Extract(doc)
	if len(got) != 1 {
		t.Fatalf("extracted %d entities, want 1", len(got))
	}
	if got[0].Type != "REGION" {
		t.Fatalf("type = %q, want REGION", got[0].Type)
	}
	if Extract(nil) != nil && len(Extract(nil)) != 0 {
		t.Fatal("nil drawing should produce no entities")
	}
}
