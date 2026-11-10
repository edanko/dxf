// Upright text normalization: readable orientation without moving the
// insertion point.
package entity

// MakeUpright normalizes the rotation of a text-like entity so its readable
// direction points rightwards (rotation within [-90, 90] degrees). The
// insertion point is preserved; rotations in (90, 270) are flipped by 180
// degrees and the MTEXT attachment corner is mirrored. It returns true when
// the entity was changed.
func MakeUpright(e Entity) bool {
	switch t := e.(type) {
	case *Text:
		r, flip := uprightAngle(t.Rotation)
		if !flip {
			return false
		}
		t.Rotation = r
		flipTextAlignment(t)
		return true
	case *MText:
		r, flip := uprightAngle(t.Rotation)
		if !flip {
			return false
		}
		t.Rotation = r
		t.AttachmentPoint = mirroredAttachment(t.AttachmentPoint)
		return true
	case interface{ GetRotation() float64; SetRotation(float64) }:
		r, flip := uprightAngle(t.GetRotation())
		if !flip {
			return false
		}
		t.SetRotation(r)
		return true
	default:
		return false
	}
}

// uprightAngle returns the normalized rotation; flip is true when the text
// reads upside down (rotation in (90, 270) after normalization).
func uprightAngle(rot float64) (normalized float64, flip bool) {
	normalized = normalizeAngleDeg(rot)
	if normalized > 90 && normalized < 270 {
		return normalizeAngleDeg(normalized - 180), true
	}
	return normalized, false
}

func normalizeAngleDeg(a float64) float64 {
	for a < 0 {
		a += 360
	}
	for a >= 360 {
		a -= 360
	}
	return a
}

// flipTextAlignment swaps the horizontal justification side so the text stays
// anchored on the same visual side after a 180-degree flip (left=1 <->
// right=2, centered unchanged).
func flipTextAlignment(t *Text) {
	switch t.HorizontalFlag {
	case 1:
		t.HorizontalFlag = 2
	case 2:
		t.HorizontalFlag = 1
	}
}

// mirroredAttachment maps an MTEXT attachment corner to its 180-degree
// opposite: TL(1)<->BR(9), TC(2)<->BC(8), TR(3)<->BL(7), ML(4)<->MR(6);
// middle-center (5) is unchanged.
func mirroredAttachment(a MTextAttachmentPoint) MTextAttachmentPoint {
	if a < MTextTopLeft || a > MTextBottomRight {
		return a
	}
	return MTextAttachmentPoint(10 - int(a))
}
