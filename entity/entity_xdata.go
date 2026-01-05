package entity

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/xdata"
)

func (e *entity) HasXData(appID string) bool {
	if e.xdata == nil {
		return false
	}
	return e.xdata.HasAppID(appID)
}

func (e *entity) GetXData(appID string) (xdata.AppData, bool) {
	if e.xdata == nil {
		return xdata.AppData{}, false
	}
	return e.xdata.GetAppData(appID)
}

func (e *entity) SetXData(appID string, tags []xdata.Tag) {
	if e.xdata == nil {
		e.xdata = xdata.NewXData()
	}
	e.xdata.SetAppData(appID, tags)
}

func (e *entity) AddXDataTag(appID string, code int, value interface{}) error {
	if e.xdata == nil {
		e.xdata = xdata.NewXData()
	}
	return e.xdata.AddTag(appID, code, value)
}

func (e *entity) RemoveXData(appID string) {
	if e.xdata != nil {
		e.xdata.RemoveAppID(appID)
	}
}

func (e *entity) ClearXData() {
	if e.xdata != nil {
		e.xdata.Clear()
	}
}

func (e *entity) XDataAppIDs() []string {
	if e.xdata == nil {
		return []string{}
	}
	return e.xdata.AppIDs()
}

func (e *entity) GetXDataObject() *xdata.XData {
	return e.xdata
}

func (e *entity) SetXDataObject(x *xdata.XData) {
	e.xdata = x
}

func (e *entity) FormatXData(f format.Formatter) {
	if e.xdata != nil && e.xdata.Len() > 0 {
		xf := xdata.NewFormatter()
		xf.FormatTo(f, e.xdata)
	}
}
