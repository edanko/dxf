package xdata

import (
	"fmt"
	"testing"
)

func TestNewXData(t *testing.T) {
	xdata := NewXData()
	if xdata == nil {
		t.Error("NewXData should return non-nil")
	}
	if xdata.Apps == nil {
		t.Error("Apps map should be initialized")
	}
}

func TestXDataHasAppID(t *testing.T) {
	xdata := NewXData()
	if xdata.HasAppID("ACAD") {
		t.Error("XData should not have ACAD initially")
	}

	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})
	if !xdata.HasAppID("ACAD") {
		t.Error("XData should have ACAD after SetAppData")
	}
}

func TestXDataGetAppData(t *testing.T) {
	xdata := NewXData()
	_, ok := xdata.GetAppData("ACAD")
	if ok {
		t.Error("Should not find non-existent app data")
	}

	tags := []Tag{{Code: 1000, Value: "test"}}
	xdata.SetAppData("ACAD", tags)
	appData, ok := xdata.GetAppData("ACAD")
	if !ok {
		t.Error("Should find ACAD app data")
	}
	if appData.AppID != "ACAD" {
		t.Errorf("Expected AppID 'ACAD', got %s", appData.AppID)
	}
	if len(appData.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(appData.Tags))
	}
}

func TestXDataSetAppData(t *testing.T) {
	xdata := NewXData()
	tags := []Tag{{Code: 1000, Value: "test"}}
	xdata.SetAppData("ACAD", tags)

	if len(xdata.Apps) != 1 {
		t.Errorf("Expected 1 app, got %d", len(xdata.Apps))
	}
}

func TestXDataAddTag(t *testing.T) {
	xdata := NewXData()

	err := xdata.AddTag("ACAD", 1000, "test")
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	if !xdata.HasAppID("ACAD") {
		t.Error("XData should have ACAD after AddTag")
	}

	appData, _ := xdata.GetAppData("ACAD")
	if len(appData.Tags) != 2 {
		t.Errorf("Expected 2 tags (including marker), got %d", len(appData.Tags))
	}

	err = xdata.AddTag("ACAD", 9999, "invalid")
	if err == nil {
		t.Error("AddTag should fail for invalid group code")
	}
}

func TestXDataRemoveAppID(t *testing.T) {
	xdata := NewXData()
	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})

	xdata.RemoveAppID("ACAD")
	if xdata.HasAppID("ACAD") {
		t.Error("XData should not have ACAD after RemoveAppID")
	}
}

func TestXDataClear(t *testing.T) {
	xdata := NewXData()
	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})
	xdata.SetAppData("OTHER", []Tag{{Code: 1000, Value: "test2"}})

	xdata.Clear()
	if xdata.Len() != 0 {
		t.Errorf("Expected 0 apps after Clear, got %d", xdata.Len())
	}
}

func TestXDataLen(t *testing.T) {
	xdata := NewXData()
	if xdata.Len() != 0 {
		t.Errorf("Expected 0 length initially, got %d", xdata.Len())
	}

	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})
	if xdata.Len() != 1 {
		t.Errorf("Expected length 1, got %d", xdata.Len())
	}
}

func TestXDataAllTags(t *testing.T) {
	xdata := NewXData()
	tags := xdata.AllTags()
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags initially, got %d", len(tags))
	}

	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test1"}, {Code: 1001, Value: "test2"}})
	xdata.SetAppData("OTHER", []Tag{{Code: 1000, Value: "test3"}})

	allTags := xdata.AllTags()
	if len(allTags) != 3 {
		t.Errorf("Expected 3 tags total, got %d", len(allTags))
	}
}

func TestXDataAppIDs(t *testing.T) {
	xdata := NewXData()
	ids := xdata.AppIDs()
	if len(ids) != 0 {
		t.Errorf("Expected 0 app IDs initially, got %d", len(ids))
	}

	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})
	xdata.SetAppData("OTHER", []Tag{{Code: 1000, Value: "test2"}})

	ids = xdata.AppIDs()
	if len(ids) != 2 {
		t.Errorf("Expected 2 app IDs, got %d", len(ids))
	}
}

func TestTagString(t *testing.T) {
	tag := Tag{Code: 1000, Value: "test"}
	s := tag.String()
	if s != "(1000 test)" {
		t.Errorf("Expected '(1000 test)', got %s", s)
	}
}

func TestAppDataString(t *testing.T) {
	appData := AppData{
		AppID: "ACAD",
		Tags:  []Tag{{Code: 1000, Value: "test"}},
	}
	s := appData.String()
	if !contains(s, "ACAD") {
		t.Errorf("Expected AppData string to contain 'ACAD', got %s", s)
	}
}

func TestXDataString(t *testing.T) {
	xdata := NewXData()
	s := xdata.String()
	if s != "XData([])" {
		t.Errorf("Expected 'XData([])', got %s", s)
	}

	xdata.SetAppData("ACAD", []Tag{{Code: 1000, Value: "test"}})
	s = xdata.String()
	if !contains(s, "ACAD") {
		t.Errorf("Expected XData string to contain 'ACAD', got %s", s)
	}
}

func TestXDataMarker(t *testing.T) {
	if XDATA_MARKER != 1001 {
		t.Errorf("Expected XDATA_MARKER 1001, got %d", XDATA_MARKER)
	}
}

func TestValidXDataGroupCodes(t *testing.T) {
	validCodes := []int{1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010, 1011, 1012, 1013, 1014, 1015, 1016, 1017, 1018, 1019, 1020, 1021, 1022, 1023, 1024, 1025, 1026, 1027, 1028, 1029, 1030, 1031, 1032, 1033, 1034, 1035, 1036, 1037, 1038, 1039, 1040, 1041, 1042, 1043, 1044, 1045, 1046, 1047, 1048, 1049, 1070, 1071}

	for _, code := range validCodes {
		if !ValidXDataGroupCodes[code] {
			t.Errorf("Code %d should be valid", code)
		}
	}

	if ValidXDataGroupCodes[999] {
		t.Error("Code 999 should not be valid")
	}
}

func TestParserNew(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Error("NewParser should not return nil")
	}
	if p.Tags == nil {
		t.Error("Tags should be initialized")
	}
	if p.XData == nil {
		t.Error("XData should be initialized")
	}
	if p.Position != 0 {
		t.Error("Position should be 0")
	}
	if len(p.Errors) != 0 {
		t.Error("Errors should be empty")
	}
}

func TestParserAddTag(t *testing.T) {
	p := NewParser()
	p.AddTag(1000, "test")

	if len(p.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(p.Tags))
	}
	if p.Tags[0].Code != 1000 {
		t.Errorf("Expected Code 1000, got %d", p.Tags[0].Code)
	}
	if p.Tags[0].Value != "test" {
		t.Errorf("Expected Value 'test', got '%v'", p.Tags[0].Value)
	}
}

func TestParserHasErrorsFalse(t *testing.T) {
	p := NewParser()
	if p.HasErrors() {
		t.Error("HasErrors should return false when no errors")
	}
}

func TestParserHasErrorsTrue(t *testing.T) {
	p := NewParser()
	p.Errors = append(p.Errors, fmt.Errorf("test error"))
	if !p.HasErrors() {
		t.Error("HasErrors should return true when errors exist")
	}
}

func TestParserParseEmpty(t *testing.T) {
	p := NewParser()
	err := p.Parse()
	if err != nil {
		t.Errorf("Parse should not fail on empty tags: %v", err)
	}
	if p.XData == nil {
		t.Error("XData should not be nil after Parse")
	}
}

func TestParserParseSingleAppID(t *testing.T) {
	p := NewParser()
	p.AddTag(XDATA_MARKER, "MYAPP")
	p.AddTag(1000, "string value")

	err := p.Parse()
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if !p.XData.HasAppID("MYAPP") {
		t.Error("XData should have MYAPP")
	}

	appData, ok := p.XData.GetAppData("MYAPP")
	if !ok {
		t.Error("Should get app data for MYAPP")
	}
	if len(appData.Tags) != 2 {
		t.Errorf("Expected 2 tags (marker + data), got %d", len(appData.Tags))
	}
}

func TestParserParseMultipleApps(t *testing.T) {
	p := NewParser()
	p.AddTag(XDATA_MARKER, "APP1")
	p.AddTag(1000, "value1")
	p.AddTag(XDATA_MARKER, "APP2")
	p.AddTag(1001, "value2")

	err := p.Parse()
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if !p.XData.HasAppID("APP1") {
		t.Error("XData should have APP1")
	}
	if !p.XData.HasAppID("APP2") {
		t.Error("XData should have APP2")
	}
}

func TestParserGetXData(t *testing.T) {
	p := NewParser()
	p.AddTag(XDATA_MARKER, "TESTAPP")
	p.AddTag(1000, "test")

	xdata := p.GetXData()
	if xdata != p.XData {
		t.Error("GetXData should return the XData object")
	}
}

func TestParseTags(t *testing.T) {
	tags := []Tag{
		{Code: XDATA_MARKER, Value: "MYAPP"},
		{Code: 1000, Value: "test"},
	}

	xdata, err := ParseTags(tags)
	if err != nil {
		t.Errorf("ParseTags failed: %v", err)
	}
	if !xdata.HasAppID("MYAPP") {
		t.Error("XData should have MYAPP")
	}
}

func TestParserInvalidGroupCode(t *testing.T) {
	p := NewParser()
	p.AddTag(XDATA_MARKER, "MYAPP")
	p.AddTag(999, "invalid")

	err := p.Parse()
	if err == nil {
		t.Error("Parse should fail with invalid group code")
	}
}

func TestParserTagBeforeMarker(t *testing.T) {
	p := NewParser()
	p.AddTag(1000, "before marker")

	err := p.Parse()
	if err == nil {
		t.Error("Parse should fail when tag comes before marker")
	}
}

func TestValidXDataGroupCodesExtended(t *testing.T) {
	if !ValidXDataGroupCodes[1071] {
		t.Error("1071 should be a valid group code")
	}
	if ValidXDataGroupCodes[9999] {
		t.Error("9999 should not be a valid group code")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
