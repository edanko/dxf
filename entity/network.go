package entity

import (
	"github.com/edanko/dxf/format"
)

type Network struct {
	*entity
	Name    string
	Flags   int
	Classes []string
}

func NewNetwork() *Network {
	return &Network{
		entity:  NewEntity(NETWORK),
		Name:    "",
		Flags:   0,
		Classes: make([]string, 0),
	}
}

func (n *Network) IsEntity() bool {
	return true
}

func (n *Network) Format(f format.Formatter) {
	n.entity.Format(f)
	f.WriteString(100, "AcDbNetwork")
	f.WriteString(2, n.Name)
	f.WriteInt(70, n.Flags)
	for _, class := range n.Classes {
		f.WriteString(100, class)
	}
}

func (n *Network) BBox() ([]float64, []float64) {
	return []float64{0, 0, 0}, []float64{0, 0, 0}
}

func (n *Network) Copy() Entity {
	net := NewNetwork()
	net.entity = n.entity
	net.Name = n.Name
	net.Flags = n.Flags
	net.Classes = make([]string, len(n.Classes))
	for i, c := range n.Classes {
		net.Classes[i] = c
	}
	return net
}

func (n *Network) Validate() error {
	return nil
}

func (n *Network) AddClass(class string) {
	n.Classes = append(n.Classes, class)
}

func (n *Network) ClearClasses() {
	n.Classes = make([]string, 0)
}
