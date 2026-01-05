package main

import (
	"log"
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
)

func main() {
	d, err := drawing.New()
	if err != nil {
		log.Fatal(err)
	}
	d.Header().LtScale = 100.0
	d.AddLayer("Toroidal", color.White, d.LtContinuous())
	d.AddLayer("Poloidal", color.Red, d.LtHidden())
	z := 0.0
	r1 := 200.0
	r2 := 500.0
	ndiv := 16
	dtheta := 2.0 * math.Pi / float64(ndiv)
	theta := 0.0
	for i := 0; i < ndiv; i++ {
		d.ChangeLayer("Toroidal")
		_, err := d.Circle(0.0, 0.0, z+r1*math.Cos(theta), r2-r1*math.Sin(theta))
		if err != nil {
			log.Printf("Error creating circle: %v", err)
		}
		d.ChangeLayer("Poloidal")
		_, err = d.Circle(r2*math.Cos(theta), r2*math.Sin(theta), 0.0, r1)
		if err != nil {
			log.Printf("Error creating circle: %v", err)
		}
		// TODO: Implement SetExtrusion function
		// dxf.SetExtrusion(c, []float64{-1.0 * math.Sin(theta), math.Cos(theta), 0.0})
		theta += dtheta
	}
	err = d.SaveAs("torus.dxf")
	if err != nil {
		log.Fatal(err)
	}
}
