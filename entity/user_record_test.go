package entity

import (
	"reflect"
	"strings"
	"testing"
)

func TestUserRecordNestedRoundTrip(t *testing.T) {
	rec := &UserRecord{
		Name: "settings",
		Fields: []UserField{
			{Key: "title", Value: UserString("hello")},
			{Key: "count", Value: UserInt(42)},
			{Key: "scale", Value: UserFloat(1.25)},
			{Key: "origin", Value: UserPoint(1, 2, 3)},
			{Key: "items", Value: UserList(
				UserString("a"),
				UserInt(7),
				UserDictionary(
					UserField{Key: "nested", Value: UserList(UserFloat(2.5), UserString("z"))},
				),
			)},
		},
	}
	tags, err := rec.Compile()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseUserRecord(tags)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Fields, rec.Fields) || got.Name != rec.Name {
		t.Fatalf("round trip mismatch:\n got %#v\nwant %#v", got, rec)
	}
	if len(got.Extra) != 0 {
		t.Fatalf("unexpected extra tags: %#v", got.Extra)
	}
}

func TestUserRecordApplyPreservesUnrelatedTags(t *testing.T) {
	x := NewXRecord()
	x.AddTag(999, "keep me")
	x.AddTag(1, "unrelated")
	rec := &UserRecord{Name: "block", Fields: []UserField{{Key: "n", Value: UserInt(1)}}}
	if err := rec.ApplyTo(x); err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseUserRecord(x.Tags)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Extra) != 2 || parsed.Extra[0].Value != "keep me" || parsed.Extra[1].Value != "unrelated" {
		t.Fatalf("unrelated tags not preserved: %#v", parsed.Extra)
	}
	// Re-applying must not duplicate the user block.
	if err := rec.ApplyTo(x); err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(strings.Join(tagValues(x.Tags), ","), "{USER"); count != 1 {
		t.Fatalf("user block duplicated: %d", count)
	}
}

func TestUserRecordRejectsInvalidStrings(t *testing.T) {
	if _, err := (&UserRecord{Name: "bad\nname"}).Compile(); err == nil {
		t.Fatal("expected newline rejection")
	}
	long := strings.Repeat("x", MaxUserRecordString+1)
	if _, err := (&UserRecord{Name: "ok", Fields: []UserField{{Key: "k", Value: UserString(long)}}}).Compile(); err == nil {
		t.Fatal("expected overlong rejection")
	}
	if _, err := ParseUserRecord([]XRecordTag{{Code: 1, Value: "no marker"}}); err == nil {
		t.Fatal("expected missing marker error")
	}
	if _, err := ParseUserRecord([]XRecordTag{
		{Code: 302, Value: "{USER"}, {Code: 1, Value: "x"}, {Code: 1, Value: "k"}, {Code: 302, Value: "{LIST"},
	}); err == nil {
		t.Fatal("expected unterminated list error")
	}
}

func tagValues(tags []XRecordTag) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if s, ok := tag.Value.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func TestBinaryRecordEveryByteRoundTrip(t *testing.T) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	rec := &BinaryRecord{Name: "payload", Data: data}
	x := NewXRecord()
	x.AddTag(999, "keep")
	if err := rec.ApplyTo(x); err != nil {
		t.Fatal(err)
	}
	got, err := ParseBinaryRecord(x.Tags)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != rec.Name || !reflect.DeepEqual(got.Data, data) {
		t.Fatalf("binary round trip mismatch: name=%q len=%d", got.Name, len(got.Data))
	}
	if len(got.Extra) != 1 || got.Extra[0].Value != "keep" {
		t.Fatalf("unrelated tag not preserved: %#v", got.Extra)
	}
}

func TestBinaryRecordEmptyAndInvalid(t *testing.T) {
	rec := &BinaryRecord{Name: "empty"}
	tags, err := rec.Compile()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseBinaryRecord(tags)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Data) != 0 {
		t.Fatalf("empty payload decoded %d bytes", len(got.Data))
	}
	if _, err := ParseBinaryRecord([]XRecordTag{{Code: 302, Value: "{BINARY"}, {Code: 1, Value: "x"}, {Code: 310, Value: "zz"}}); err == nil {
		t.Fatal("expected invalid hex error")
	}
}
