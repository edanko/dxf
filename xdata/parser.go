package xdata

import (
	"fmt"
	"strings"
)

type Parser struct {
	Tags     []Tag
	Position int
	Errors   []error
	XData    *XData
}

func NewParser() *Parser {
	return &Parser{
		Tags:  make([]Tag, 0),
		XData: NewXData(),
	}
}

func (p *Parser) AddTag(code int, value interface{}) {
	p.Tags = append(p.Tags, Tag{Code: code, Value: value})
}

func (p *Parser) HasErrors() bool {
	return len(p.Errors) > 0
}

func (p *Parser) Parse() error {
	p.XData = NewXData()
	var currentAppID string
	var currentTags []Tag

	for i, tag := range p.Tags {
		if tag.Code == XDATA_MARKER {
			if currentAppID != "" && len(currentTags) > 0 {
				p.XData.SetAppData(currentAppID, currentTags)
			}
			currentAppID = tag.Value.(string)
			currentTags = []Tag{tag}
		} else if currentAppID != "" {
			if !ValidXDataGroupCodes[tag.Code] {
				p.Errors = append(p.Errors, fmt.Errorf("invalid XDATA group code %d at position %d", tag.Code, i))
				continue
			}
			currentTags = append(currentTags, tag)
		} else {
			p.Errors = append(p.Errors, fmt.Errorf("XDATA tag %d found before APPID marker", tag.Code))
		}
	}

	if currentAppID != "" && len(currentTags) > 0 {
		p.XData.SetAppData(currentAppID, currentTags)
	}

	if len(p.Errors) > 0 {
		errMsgs := make([]string, len(p.Errors))
		for i, err := range p.Errors {
			errMsgs[i] = err.Error()
		}
		return fmt.Errorf("XDATA parsing errors: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

func (p *Parser) GetXData() *XData {
	return p.XData
}

func ParseTags(tags []Tag) (*XData, error) {
	parser := NewParser()
	parser.Tags = tags
	err := parser.Parse()
	return parser.XData, err
}
