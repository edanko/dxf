package dxf

import (
	"bytes"
	"testing"

	"github.com/edanko/dxf/object"
)

func TestMLeaderStyleCollectionFromRootDictionary(t *testing.T) {
	d, err := FromFile("testdata/3dface.dxf")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	root, ok := d.RootDictionary()
	if !ok {
		t.Fatal("no root dictionary")
	}
	obj, ok := d.GetObjectByDictionaryKey("ACAD_MLEADERSTYLE")
	if !ok {
		t.Fatal("ACAD_MLEADERSTYLE not resolved by key")
	}
	collection, ok := obj.(*object.MLeaderStyleCollection)
	if !ok {
		t.Fatalf("type = %T, want *object.MLeaderStyleCollection", obj)
	}
	handle := collection.Handle()
	owner := collection.Owner
	if handle == "" || owner != root.Handle() {
		t.Fatalf("handle/owner = %q/%q, root = %q", handle, owner, root.Handle())
	}
	byHandle, ok := d.GetObject(handle)
	if !ok {
		t.Fatalf("collection handle %q not found", handle)
	}
	if _, ok := byHandle.(*object.MLeaderStyleCollection); !ok {
		t.Fatalf("by-handle type = %T", byHandle)
	}
	handles := collection.MemberHandles()
	if len(handles) == 0 {
		t.Fatal("no member handles")
	}
	for _, memberHandle := range handles {
		if _, ok := d.GetObject(memberHandle); !ok {
			t.Errorf("member %q does not resolve", memberHandle)
		}
	}

	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	obj, ok = reloaded.GetObjectByDictionaryKey("ACAD_MLEADERSTYLE")
	if !ok {
		t.Fatal("ACAD_MLEADERSTYLE missing after reload")
	}
	reloadedCollection, ok := obj.(*object.MLeaderStyleCollection)
	if !ok {
		t.Fatalf("reloaded type = %T", obj)
	}
	if reloadedCollection.Handle() != handle || reloadedCollection.Owner != owner {
		t.Fatalf("reloaded handle/owner = %q/%q", reloadedCollection.Handle(), reloadedCollection.Owner)
	}
	if len(reloadedCollection.MemberHandles()) != len(handles) {
		t.Fatalf("reloaded members = %v, want %v", reloadedCollection.MemberHandles(), handles)
	}
}

func TestMLineStyleCollectionFromRootDictionary(t *testing.T) {
	d, err := FromFile("testdata/3dface.dxf")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	obj, ok := d.GetObjectByDictionaryKey("ACAD_MLINESTYLE")
	if !ok {
		t.Fatal("ACAD_MLINESTYLE not resolved by key")
	}
	collection, ok := obj.(*object.MLineStyleCollection)
	if !ok {
		t.Fatalf("type = %T, want *object.MLineStyleCollection", obj)
	}
	if len(collection.MemberHandles()) == 0 {
		t.Fatal("no member handles")
	}
}
