// Revision cloud creation.
package drawing

import (
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/query"
)

// AddRevCloud adds a closed revision-cloud LWPOLYLINE around the rectangle
// (minX,minY)-(maxX,maxY). arcLength selects the arc chord size
// (query.DefaultRevCloudArcLength when <= 0) and bulge the arc style
// (query.DefaultRevCloudBulge for semicircles when 0).
func (s *Space) AddRevCloud(minX, minY, maxX, maxY, arcLength, bulge float64, opts ...EntityOption) (*entity.LwPolyline, error) {
	p, err := query.RevCloudRectangle(minX, minY, maxX, maxY, arcLength, bulge)
	if err != nil {
		return nil, err
	}
	if err := s.Add(p, opts...); err != nil {
		return nil, err
	}
	return p, nil
}

// AddRevCloud adds the revision cloud to model space.
func (d *Drawing) AddRevCloud(minX, minY, maxX, maxY, arcLength, bulge float64, opts ...EntityOption) (*entity.LwPolyline, error) {
	return d.Model().AddRevCloud(minX, minY, maxX, maxY, arcLength, bulge, opts...)
}
