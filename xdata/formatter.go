package xdata

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
)

type Formatter struct {
	Tags   []Tag
	Errors []error
}

func NewFormatter() *Formatter {
	return &Formatter{
		Tags:   make([]Tag, 0),
		Errors: make([]error, 0),
	}
}

func (f *Formatter) WriteTag(code int, value interface{}) {
	if !ValidXDataGroupCodes[code] {
		f.Errors = append(f.Errors, fmt.Errorf("invalid XDATA group code: %d", code))
		return
	}
	f.Tags = append(f.Tags, Tag{Code: code, Value: value})
}

func (f *Formatter) WriteString(code int, value string) {
	if code >= 1000 && code <= 1009 || code >= 1025 && code <= 1071 {
		f.Tags = append(f.Tags, Tag{Code: code, Value: value})
	} else {
		f.Errors = append(f.Errors, fmt.Errorf("group code %d not valid for string", code))
	}
}

func (f *Formatter) WriteFloat(code int, value float64) {
	if code >= 1040 && code <= 1048 {
		f.Tags = append(f.Tags, Tag{Code: code, Value: value})
	} else {
		f.Errors = append(f.Errors, fmt.Errorf("group code %d not valid for float", code))
	}
}

func (f *Formatter) WriteInt(code int, value int) {
	if code >= 1070 && code <= 1071 || code >= 1000 && code <= 1009 {
		f.Tags = append(f.Tags, Tag{Code: code, Value: value})
	} else if code >= 1010 && code <= 1019 {
		f.Tags = append(f.Tags, Tag{Code: code, Value: value})
	} else {
		f.Errors = append(f.Errors, fmt.Errorf("group code %d not valid for int", code))
	}
}

func (f *Formatter) Output() string {
	var lines []string
	for _, tag := range f.Tags {
		lines = append(lines, fmt.Sprintf("%d\n%v\n", tag.Code, tag.Value))
	}
	return strings.Join(lines, "")
}

func (f *Formatter) TagsFromXData(x *XData) {
	for _, appData := range x.Apps {
		for _, tag := range appData.Tags {
			f.WriteTag(tag.Code, tag.Value)
		}
	}
}

func (f *Formatter) FormatTo(fm format.Formatter, x *XData) {
	for _, appData := range x.Apps {
		for _, tag := range appData.Tags {
			switch v := tag.Value.(type) {
			case string:
				fm.WriteString(tag.Code, v)
			case float64:
				fm.WriteFloat(tag.Code, v)
			case int:
				fm.WriteInt(tag.Code, v)
			default:
				fm.WriteString(tag.Code, fmt.Sprintf("%v", v))
			}
		}
	}
}
