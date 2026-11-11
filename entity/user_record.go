// User records: typed values stored in XRECORD tags. The encoding is
// recursive and self-delimiting so nested lists and dictionaries survive a
// compile/parse round trip. Tags outside the user-record block are preserved
// when the helper is applied to an existing XRECORD.
package entity

import (
	"fmt"
	"math"
	"strings"
)

// MaxUserRecordString is the maximum length of a user-record string or
// dictionary key. Newlines are rejected because DXF string tags are
// line-oriented.
const MaxUserRecordString = 2048

// UserValueKind identifies the type of a UserValue.
type UserValueKind int

const (
	UserStringKind UserValueKind = iota
	UserIntKind
	UserFloatKind
	UserPointKind
	UserListKind
	UserDictionaryKind
)

// UserField is one named value of a user record.
type UserField struct {
	Key   string
	Value UserValue
}

// UserValue is one typed user-record value.
type UserValue struct {
	Kind   UserValueKind
	String string
	Int    int
	Float  float64
	Point  []float64
	List   []UserValue
	Dict   []UserField
}

// UserString creates a string value.
func UserString(s string) UserValue { return UserValue{Kind: UserStringKind, String: s} }

// UserInt creates an integer value.
func UserInt(i int) UserValue { return UserValue{Kind: UserIntKind, Int: i} }

// UserFloat creates a float value.
func UserFloat(f float64) UserValue { return UserValue{Kind: UserFloatKind, Float: f} }

// UserPoint creates a point value. The slice is copied.
func UserPoint(p ...float64) UserValue {
	return UserValue{Kind: UserPointKind, Point: append([]float64(nil), p...)}
}

// UserList creates a list value.
func UserList(values ...UserValue) UserValue {
	return UserValue{Kind: UserListKind, List: values}
}

// UserDictionary creates a dictionary value.
func UserDictionary(fields ...UserField) UserValue {
	return UserValue{Kind: UserDictionaryKind, Dict: fields}
}

// UserRecord is a named collection of typed fields.
type UserRecord struct {
	Name   string
	Fields []UserField
	// Extra holds XRECORD tags outside the user-record block so applying the
	// record to an existing entity can preserve unrelated data.
	Extra []XRecordTag
}

const (
	userRecordOpen  = "{USER"
	userRecordClose = "}"
	userListOpen    = "{LIST"
	userDictOpen    = "{DICT"
)

// Compile encodes the record as XRECORD tags. The record is wrapped in code
// 302 structure markers; strings use code 1, integers 90, floats 40, and
// points 10/20/30.
func (r *UserRecord) Compile() ([]XRecordTag, error) {
	if err := validateUserString("record name", r.Name); err != nil {
		return nil, err
	}
	tags := []XRecordTag{{Code: 302, Value: userRecordOpen}, {Code: 1, Value: r.Name}}
	for _, field := range r.Fields {
		if err := validateUserString("field key", field.Key); err != nil {
			return nil, err
		}
		valueTags, err := compileUserValue(field.Value)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.Key, err)
		}
		tags = append(tags, XRecordTag{Code: 1, Value: field.Key})
		tags = append(tags, valueTags...)
	}
	tags = append(tags, XRecordTag{Code: 302, Value: userRecordClose})
	return tags, nil
}

// ApplyTo replaces the user-record block in x, leaving unrelated tags intact.
func (r *UserRecord) ApplyTo(x *XRecord) error {
	if x == nil {
		return fmt.Errorf("user record: nil XRECORD")
	}
	compiled, err := r.Compile()
	if err != nil {
		return err
	}
	kept := stripUserRecordBlock(x.Tags)
	x.Tags = append(kept, compiled...)
	return nil
}

func compileUserValue(v UserValue) ([]XRecordTag, error) {
	switch v.Kind {
	case UserStringKind:
		if err := validateUserString("string", v.String); err != nil {
			return nil, err
		}
		return []XRecordTag{{Code: 1, Value: v.String}}, nil
	case UserIntKind:
		return []XRecordTag{{Code: 90, Value: v.Int}}, nil
	case UserFloatKind:
		if math.IsNaN(v.Float) || math.IsInf(v.Float, 0) {
			return nil, fmt.Errorf("non-finite float")
		}
		return []XRecordTag{{Code: 40, Value: v.Float}}, nil
	case UserPointKind:
		if len(v.Point) != 2 && len(v.Point) != 3 {
			return nil, fmt.Errorf("point has %d coordinates, want 2 or 3", len(v.Point))
		}
		for _, c := range v.Point {
			if math.IsNaN(c) || math.IsInf(c, 0) {
				return nil, fmt.Errorf("non-finite point coordinate")
			}
		}
		tags := []XRecordTag{{Code: 10, Value: v.Point[0]}, {Code: 20, Value: v.Point[1]}}
		if len(v.Point) == 3 {
			tags = append(tags, XRecordTag{Code: 30, Value: v.Point[2]})
		}
		return tags, nil
	case UserListKind:
		tags := []XRecordTag{{Code: 302, Value: userListOpen}}
		for _, item := range v.List {
			itemTags, err := compileUserValue(item)
			if err != nil {
				return nil, err
			}
			tags = append(tags, itemTags...)
		}
		return append(tags, XRecordTag{Code: 302, Value: userRecordClose}), nil
	case UserDictionaryKind:
		tags := []XRecordTag{{Code: 302, Value: userDictOpen}}
		for _, field := range v.Dict {
			if err := validateUserString("dictionary key", field.Key); err != nil {
				return nil, err
			}
			valueTags, err := compileUserValue(field.Value)
			if err != nil {
				return nil, fmt.Errorf("dictionary key %q: %w", field.Key, err)
			}
			tags = append(tags, XRecordTag{Code: 1, Value: field.Key})
			tags = append(tags, valueTags...)
		}
		return append(tags, XRecordTag{Code: 302, Value: userRecordClose}), nil
	default:
		return nil, fmt.Errorf("unknown user value kind %d", v.Kind)
	}
}

// ParseUserRecord decodes the user-record block from XRECORD tags. Tags
// outside the block are retained in Extra. A missing or malformed block is an
// error.
func ParseUserRecord(tags []XRecordTag) (*UserRecord, error) {
	start := -1
	for i, tag := range tags {
		if tag.Code == 302 && stringValue(tag.Value) == userRecordOpen {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("user record: missing %q marker", userRecordOpen)
	}
	p := &userTagParser{tags: tags, pos: start + 1}
	if p.done() || p.tags[p.pos].Code != 1 {
		return nil, fmt.Errorf("user record: missing record name")
	}
	name := stringValue(p.tags[p.pos].Value)
	p.pos++
	rec := &UserRecord{Name: name}
	for !p.done() {
		key := stringValue(p.tags[p.pos].Value)
		if p.tags[p.pos].Code == 302 && key == userRecordClose {
			p.pos++
			rec.Extra = append(append([]XRecordTag(nil), tags[:start]...), tags[p.pos:]...)
			return rec, nil
		}
		if p.tags[p.pos].Code != 1 {
			return nil, fmt.Errorf("user record: expected field key, got code %d", p.tags[p.pos].Code)
		}
		p.pos++
		value, err := p.parseValue()
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", key, err)
		}
		rec.Fields = append(rec.Fields, UserField{Key: key, Value: value})
	}
	return nil, fmt.Errorf("user record: missing %q terminator", userRecordClose)
}

type userTagParser struct {
	tags []XRecordTag
	pos  int
}

func (p *userTagParser) done() bool { return p.pos >= len(p.tags) }

func (p *userTagParser) parseValue() (UserValue, error) {
	if p.done() {
		return UserValue{}, fmt.Errorf("unexpected end of tags")
	}
	tag := p.tags[p.pos]
	switch tag.Code {
	case 1:
		p.pos++
		return UserString(stringValue(tag.Value)), nil
	case 90:
		p.pos++
		return UserInt(intValue(tag.Value)), nil
	case 40:
		p.pos++
		return UserFloat(floatValue(tag.Value)), nil
	case 10:
		p.pos++
		x := floatValue(tag.Value)
		if p.done() || p.tags[p.pos].Code != 20 {
			return UserValue{}, fmt.Errorf("point missing code 20")
		}
		y := floatValue(p.tags[p.pos].Value)
		p.pos++
		point := []float64{x, y}
		if !p.done() && p.tags[p.pos].Code == 30 {
			point = append(point, floatValue(p.tags[p.pos].Value))
			p.pos++
		}
		return UserPoint(point...), nil
	case 302:
		switch stringValue(tag.Value) {
		case userListOpen:
			p.pos++
			list := UserValue{Kind: UserListKind}
			for {
				if p.done() {
					return UserValue{}, fmt.Errorf("unterminated list")
				}
				if p.tags[p.pos].Code == 302 && stringValue(p.tags[p.pos].Value) == userRecordClose {
					p.pos++
					return list, nil
				}
				item, err := p.parseValue()
				if err != nil {
					return UserValue{}, err
				}
				list.List = append(list.List, item)
			}
		case userDictOpen:
			p.pos++
			dict := UserValue{Kind: UserDictionaryKind}
			for {
				if p.done() {
					return UserValue{}, fmt.Errorf("unterminated dictionary")
				}
				if p.tags[p.pos].Code == 302 && stringValue(p.tags[p.pos].Value) == userRecordClose {
					p.pos++
					return dict, nil
				}
				if p.tags[p.pos].Code != 1 {
					return UserValue{}, fmt.Errorf("dictionary key must be a string")
				}
				key := stringValue(p.tags[p.pos].Value)
				p.pos++
				value, err := p.parseValue()
				if err != nil {
					return UserValue{}, fmt.Errorf("dictionary key %q: %w", key, err)
				}
				dict.Dict = append(dict.Dict, UserField{Key: key, Value: value})
			}
		}
	}
	return UserValue{}, fmt.Errorf("unsupported value code %d", tag.Code)
}

func validateUserString(what, s string) error {
	if len(s) > MaxUserRecordString {
		return fmt.Errorf("%s exceeds %d characters", what, MaxUserRecordString)
	}
	if strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("%s contains a newline", what)
	}
	return nil
}

func stripUserRecordBlock(tags []XRecordTag) []XRecordTag {
	start := -1
	for i, tag := range tags {
		if tag.Code == 302 && stringValue(tag.Value) == userRecordOpen {
			start = i
			break
		}
	}
	if start < 0 {
		return append([]XRecordTag(nil), tags...)
	}
	depth := 0
	for i := start; i < len(tags); i++ {
		if tags[i].Code != 302 {
			continue
		}
		switch stringValue(tags[i].Value) {
		case userRecordOpen, userListOpen, userDictOpen:
			depth++
		case userRecordClose:
			depth--
			if depth == 0 {
				return append(append([]XRecordTag(nil), tags[:start]...), tags[i+1:]...)
			}
		}
	}
	return append([]XRecordTag(nil), tags[:start]...)
}

func stringValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	}
	return fmt.Sprint(v)
}

func intValue(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int8:
		return int(x)
	case int16:
		return int(x)
	case int32:
		return int(x)
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		var n int
		fmt.Sscanf(x, "%d", &n)
		return n
	}
	return 0
}

func floatValue(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		var f float64
		fmt.Sscanf(x, "%g", &f)
		return f
	}
	return 0
}
