package models

type Scene struct {
	Materials []Material
	Spheres   []Sphere
}

func (scene *Scene) LinkMaterials() {
	// TODO: Generalize
	// TODO: Need for a lookup?
	for _, sphere := range scene.Spheres {
		for _, material := range scene.Materials {
			if sphere.MaterialID == material.ID {
				sphere.Material = &material
			}
		}
	}
}
