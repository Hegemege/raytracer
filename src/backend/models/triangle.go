package models

import (
	"math"
	"raytracer/utility"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

type Triangle struct {
	Index         int
	Vertices      [3]mgl32.Vec3
	TextureCoords [3]mgl32.Vec2
	// TODO: Add vertex normals

	// For Woop intersection
	//LocalM mgl32.Mat3
	//LocalN mgl32.Vec3

	Normal   mgl32.Vec3
	Material *gwob.Material

	// Helpers for trace
	Edge0 mgl32.Vec3
	Edge1 mgl32.Vec3
	Edge2 mgl32.Vec3

	IsLight bool
}

func (t *Triangle) Center() mgl32.Vec3 {
	return t.Vertices[0].Add(t.Vertices[1]).Add(t.Vertices[2]).Mul(1.0 / 3.0)
}

func (t *Triangle) Min() mgl32.Vec3 {
	return utility.Vec3Min(utility.Vec3Min(t.Vertices[0], t.Vertices[1]), t.Vertices[2])
}

func (t *Triangle) Max() mgl32.Vec3 {
	return utility.Vec3Max(utility.Vec3Max(t.Vertices[0], t.Vertices[1]), t.Vertices[2])
}

func NewTriangle(v0 mgl32.Vec3, v1 mgl32.Vec3, v2 mgl32.Vec3, material *gwob.Material, index int) *Triangle {
	normal := v1.Sub(v0).Cross(v2.Sub(v0)).Normalize()

	tri := &Triangle{
		Index: index,
		Vertices: [3]mgl32.Vec3{
			v0, v1, v2,
		},
		Normal:   normal,
		Material: material,
		Edge0:    v1.Sub(v0),
		Edge1:    v2.Sub(v1),
		Edge2:    v0.Sub(v2),
		IsLight:  material.Name == "Light",
		//LocalM:   mgl32.Mat3FromCols(v1.Sub(v0), v2.Sub(v0), normal).Inv(),
	}

	//tri.LocalN = tri.LocalM.Mul(-1).Mul3x1(v0)

	return tri
}

func TriangleSorter(axis mgl32.Vec3, triangles []*Triangle, startIndex int, endIndex int) {
	sort.Slice(triangles[startIndex:endIndex+1], func(i, j int) bool {
		a := axis.Dot(triangles[startIndex+i].Center())
		b := axis.Dot(triangles[startIndex+j].Center())
		if a == b {
			return triangles[startIndex+i].Index < triangles[startIndex+j].Index
		}
		return a < b
	})
}

func (triangle *Triangle) RayIntersect(ray *Ray) (float32, float32, float32) {
	// From https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection
	v0v2 := triangle.Edge2.Mul(-1)
	pvec := ray.Direction.Cross(v0v2)
	det := triangle.Edge0.Dot(pvec)
	if det < 0.0001 {
		return -1, 0, 0
	}

	invdet := 1.0 / det
	tvec := ray.Origin.Sub(triangle.Vertices[0])
	u := tvec.Dot(pvec) * invdet
	if u < 0 || u > 1 {
		return -1, 0, 0
	}

	qvec := tvec.Cross(triangle.Edge0)
	v := ray.Direction.Dot(qvec) * invdet
	if v < 0 || u+v > 1 {
		return -1, 0, 0
	}

	t := v0v2.Dot(qvec) * invdet

	return t, u, v
}

func (triangle *Triangle) GetShortestEdge() mgl32.Vec3 {
	if triangle.Edge0.Len() < triangle.Edge1.Len() {
		if triangle.Edge0.Len() < triangle.Edge2.Len() {
			return triangle.Edge0
		}
		return triangle.Edge2
	}

	if triangle.Edge1.Len() < triangle.Edge2.Len() {
		return triangle.Edge1
	}

	return triangle.Edge2
}

func (triangle *Triangle) GetMiddleEdge() mgl32.Vec3 {
	if triangle.Edge0.Len() < triangle.Edge1.Len() {
		if triangle.Edge1.Len() < triangle.Edge2.Len() {
			return triangle.Edge1
		}
		if triangle.Edge0.Len() < triangle.Edge2.Len() {
			return triangle.Edge2
		}
		return triangle.Edge0
	}

	if triangle.Edge2.Len() < triangle.Edge1.Len() {
		return triangle.Edge1
	}
	if triangle.Edge0.Len() < triangle.Edge2.Len() {
		return triangle.Edge0
	}
	return triangle.Edge2
}

func GetTriangleBounds(triangles []*Triangle) (mgl32.Vec3, mgl32.Vec3) {
	var minx, miny, minz float32 = math.MaxFloat32, math.MaxFloat32, math.MaxFloat32
	var maxx, maxy, maxz float32 = -math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32

	for _, triangle := range triangles {
		for _, vertex := range triangle.Vertices {
			if vertex.X() < minx {
				minx = vertex.X()
			}
			if vertex.Y() < miny {
				miny = vertex.Y()
			}
			if vertex.Z() < minz {
				minz = vertex.Z()
			}
			if vertex.X() > maxx {
				maxx = vertex.X()
			}
			if vertex.Y() > maxy {
				maxy = vertex.Y()
			}
			if vertex.Z() > maxz {
				maxz = vertex.Z()
			}
		}
	}

	return mgl32.Vec3{minx, miny, minz}, mgl32.Vec3{maxx, maxy, maxz}
}

/*
func NewTriangle(v0 mgl32.Vec3, v1 mgl32.Vec3, v2 mgl32.Vec3) *Triangle {
	// Woop04
	normal := v1.Sub(v0).Cross(v2.Sub(v0)).Normalize()
	tri := &Triangle{
		Vertices: [3]mgl32.Vec3{v0, v1, v2},
		LocalM:   mgl32.Mat3FromCols(v1.Sub(v0), v2.Sub(v0), normal).Inv(),
		Normal:   normal,
	}

	tri.LocalN = tri.LocalM.Mul(-1).Mul3x1(v0)

	return tri
}
*/

/*
func (triangle *Triangle) RayIntersect(ray *Ray) (float32, float32, float32) {
	// Woop04
	transformed_origin := triangle.LocalM.Mul3x1(ray.Origin).Add(triangle.LocalN)
	transformed_dir := triangle.LocalM.Mul3x1(ray.Direction)

	t := -transformed_origin.Z() / transformed_dir.Z()
	u := transformed_origin.X() + transformed_dir.X()*t
	v := transformed_origin.Y() + transformed_dir.Y()*t

	if u <= 0.0 || v <= 0.0 || u+v >= 1.0 {
		return -1, -1, -1
	}

	return t, u, v
}
*/
