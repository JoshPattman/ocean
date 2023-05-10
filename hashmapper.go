package main

import (
	"math"

	"github.com/faiface/pixel"
)

type HashMappable interface {
	HMPos() pixel.Vec
	Eq(o HashMappable) bool
}

type HashMap[T HashMappable] struct {
	Objects   []T
	areas     map[pixel.Vec][]T
	areaScale float64
}

func NewHashMap[T HashMappable](areaScale float64) *HashMap[T] {
	return &HashMap[T]{
		Objects:   make([]T, 0),
		areas:     make(map[pixel.Vec][]T),
		areaScale: areaScale,
	}
}

func (m *HashMap[T]) Add(o T) {
	m.Objects = append(m.Objects, o)
}

// Instantly removes the object from the hashmap. It will not be returned by Query anymore.
func (m *HashMap[T]) Remove(o T) {
	for i, o2 := range m.Objects {
		if o.Eq(o2) {
			m.Objects = append(m.Objects[:i], m.Objects[i+1:]...)
			break
		}
	}
	// Remove the object from the areas
	ap := m.toAreaPos(o.HMPos())
	if area, in := m.areas[ap]; in {
		for i, o2 := range area {
			if o.Eq(o2) {
				m.areas[ap] = append(area[:i], area[i+1:]...)
				break
			}
		}
	}

}

func (m *HashMap[T]) Refresh() {
	// Clear the areas
	for k := range m.areas {
		delete(m.areas, k)
	}
	// Add the objects to the areas
	for _, o := range m.Objects {
		ap := m.toAreaPos(o.HMPos())
		if area, in := m.areas[ap]; in {
			m.areas[ap] = append(area, o)
		} else {
			m.areas[ap] = []T{o}
		}
	}
}

func (m *HashMap[T]) toAreaPos(pos pixel.Vec) pixel.Vec {
	ax, ay := int(math.Round(pos.X/m.areaScale)), int(math.Round(pos.Y/m.areaScale))
	return pixel.V(float64(ax), float64(ay))
}

func (m *HashMap[T]) Query(pos pixel.Vec, radius float64) []T {
	searchAreasRadius := int(math.Ceil(radius / m.areaScale))
	ap := m.toAreaPos(pos)
	objects := make([]T, 0)
	for xo := -searchAreasRadius; xo <= searchAreasRadius; xo++ {
		for yo := -searchAreasRadius; yo <= searchAreasRadius; yo++ {
			if area, in := m.areas[ap.Add(pixel.V(float64(xo), float64(yo)))]; in {
				for _, o := range area {
					if o.HMPos().Sub(pos).Len() < radius {
						objects = append(objects, o)
					}
				}
			}
		}
	}
	return objects
}
