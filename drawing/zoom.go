// View fitting: zoom factors and view centers for extents in a viewport.
package drawing

// ViewFit describes a viewport fitted around drawing extents.
type ViewFit struct {
	CenterX float64 // view center (model coordinates)
	CenterY float64
	Width   float64 // visible width in drawing units
	Height  float64 // visible height in drawing units
	Zoom    float64 // drawing-units per viewport unit (0 when empty)
}

// FitExtentsToViewport fits (minX,minY)-(maxX,maxY) into a viewport with the
// given aspect ratio (width/height) and a proportional margin (0.1 = 10%
// around the extents). The resulting view always contains the full extents:
// the limiting dimension gets the margin, the other grows to preserve the
// aspect ratio.
func FitExtentsToViewport(minX, minY, maxX, maxY, aspect, margin float64) ViewFit {
	if aspect <= 0 {
		aspect = 1
	}
	if margin < 0 {
		margin = 0
	}
	w := (maxX - minX) * (1 + margin)
	h := (maxY - minY) * (1 + margin)
	// Grow the non-limiting dimension to satisfy the aspect ratio.
	if w/h < aspect {
		w = h * aspect
	} else {
		h = w / aspect
	}
	v := ViewFit{
		CenterX: (minX + maxX) / 2,
		CenterY: (minY + maxY) / 2,
		Width:   w,
		Height:  h,
	}
	diag := max(maxX-minX, maxY-minY)
	if diag > 0 {
		v.Zoom = diag / max(w, h)
	}
	return v
}

// FitView fits the model-space extents into a viewport of the given aspect
// ratio with margin, returning the view center, size and zoom factor.
func (d *Drawing) FitView(aspect, margin float64) (ViewFit, bool) {
	min, max, ok := d.Extents()
	if !ok {
		return ViewFit{}, false
	}
	return FitExtentsToViewport(min[0], min[1], max[0], max[1], aspect, margin), true
}

// FitLayoutView fits one layout's extents the same way.
func (d *Drawing) FitLayoutView(layout string, aspect, margin float64) (ViewFit, bool) {
	min, max, ok := d.LayoutExtents(layout)
	if !ok {
		return ViewFit{}, false
	}
	return FitExtentsToViewport(min[0], min[1], max[0], max[1], aspect, margin), true
}
