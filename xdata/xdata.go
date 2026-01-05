package xdata

import (
	"fmt"
	"strings"
)

const XDATA_MARKER = 1001

var ValidXDataGroupCodes = map[int]bool{
	1000: true, 1001: true, 1002: true, 1003: true, 1004: true,
	1005: true, 1006: true, 1007: true, 1008: true, 1009: true,
	1010: true, 1011: true, 1012: true, 1013: true, 1014: true,
	1015: true, 1016: true, 1017: true, 1018: true, 1019: true,
	1020: true, 1021: true, 1022: true, 1023: true, 1024: true,
	1025: true, 1026: true, 1027: true, 1028: true, 1029: true,
	1030: true, 1031: true, 1032: true, 1033: true, 1034: true,
	1035: true, 1036: true, 1037: true, 1038: true, 1039: true,
	1040: true, 1041: true, 1042: true, 1043: true, 1044: true,
	1045: true, 1046: true, 1047: true, 1048: true, 1049: true,
	1070: true, 1071: true,
}

type Tag struct {
	Code  int
	Value interface{}
}

func (t Tag) String() string {
	return fmt.Sprintf("(%d %v)", t.Code, t.Value)
}

type AppData struct {
	AppID string
	Tags  []Tag
}

func (a AppData) String() string {
	tags := make([]string, len(a.Tags))
	for i, tag := range a.Tags {
		tags[i] = tag.String()
	}
	return fmt.Sprintf("AppData(%s: %v)", a.AppID, strings.Join(tags, ", "))
}

type XData struct {
	Apps map[string]AppData
}

func NewXData() *XData {
	return &XData{
		Apps: make(map[string]AppData),
	}
}

func (x *XData) HasAppID(appID string) bool {
	_, ok := x.Apps[appID]
	return ok
}

func (x *XData) GetAppData(appID string) (AppData, bool) {
	app, ok := x.Apps[appID]
	return app, ok
}

func (x *XData) SetAppData(appID string, tags []Tag) {
	x.Apps[appID] = AppData{
		AppID: appID,
		Tags:  tags,
	}
}

func (x *XData) AddTag(appID string, code int, value interface{}) error {
	if !ValidXDataGroupCodes[code] {
		return fmt.Errorf("invalid XDATA group code: %d", code)
	}

	app, ok := x.Apps[appID]
	if !ok {
		app = AppData{
			AppID: appID,
			Tags:  []Tag{{Code: XDATA_MARKER, Value: appID}},
		}
	}
	app.Tags = append(app.Tags, Tag{Code: code, Value: value})
	x.Apps[appID] = app
	return nil
}

func (x *XData) RemoveAppID(appID string) {
	delete(x.Apps, appID)
}

func (x *XData) Clear() {
	x.Apps = make(map[string]AppData)
}

func (x *XData) Len() int {
	return len(x.Apps)
}

func (x *XData) AllTags() []Tag {
	var allTags []Tag
	for _, app := range x.Apps {
		allTags = append(allTags, app.Tags...)
	}
	return allTags
}

func (x *XData) AppIDs() []string {
	ids := make([]string, 0, len(x.Apps))
	for id := range x.Apps {
		ids = append(ids, id)
	}
	return ids
}

func (x *XData) String() string {
	appIDs := make([]string, 0, len(x.Apps))
	for id := range x.Apps {
		appIDs = append(appIDs, id)
	}
	return fmt.Sprintf("XData(%v)", appIDs)
}
