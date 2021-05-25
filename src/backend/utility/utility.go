package utility

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/udhos/gwob"
)

func ProgressUpdate(progress float32, event string, taskId int, rays uint64) {
	data := struct {
		Progress float32 `json:"progress"`
		Event    string  `json:"event"`
		TaskID   int     `json:"taskId"`
		Rays     uint64  `json:"rays"`
	}{
		progress,
		event,
		taskId,
		rays,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	js.Global().Call("progressUpdate", string(raw))
}

func ClampColor(c mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Min(math.Max(float64(c.X()), 0), 1.0)),
		float32(math.Min(math.Max(float64(c.Y()), 0), 1.0)),
		float32(math.Min(math.Max(float64(c.Z()), 0), 1.0)),
	}
}

func MultiplyColor(c1 mgl32.Vec3, c2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		c1[0] * c2[0],
		c1[1] * c2[1],
		c1[2] * c2[2],
	}
}

func RandomInHemisphere(normal mgl32.Vec3) mgl32.Vec3 {
	inUnitSphere := RandomInUnitSphere()
	if inUnitSphere.Dot(normal) > 0.0 {
		return inUnitSphere
	}

	return inUnitSphere.Mul(-1)
}

func RandomInUnitSphere() mgl32.Vec3 {
	for {
		p := mgl32.Vec3{
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
			rand.Float32()*2 - 1,
		}
		if p.LenSqr() < 1 {
			return p
		}
	}
}

func BoolToInt(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// Returns combined minimum of the two vectors
func Vec3Min(v1 mgl32.Vec3, v2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Min(float64(v1.X()), float64(v2.X()))),
		float32(math.Min(float64(v1.Y()), float64(v2.Y()))),
		float32(math.Min(float64(v1.Z()), float64(v2.Z()))),
	}
}

// Returns combined maximum of the two vectors
func Vec3Max(v1 mgl32.Vec3, v2 mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(math.Max(float64(v1.X()), float64(v2.X()))),
		float32(math.Max(float64(v1.Y()), float64(v2.Y()))),
		float32(math.Max(float64(v1.Z()), float64(v2.Z()))),
	}
}

// TextureCoordinates gets texture coordinates for a stride index.
func TextureCoordinates(o *gwob.Obj, stride int) (float32, float32, error) {
	offset := o.StrideOffsetTexture / 4
	floatsPerStride := o.StrideSize / 4
	f := offset + stride*floatsPerStride
	max := len(o.Coord)
	if f > max-2 {
		return 0, 0, errors.New("overflowing vertex coord index")
	}
	return o.Coord[f], o.Coord[f+1], nil
}

// VertexCoordinates gets vertex coordinates for a stride index.
func VertexCoordinates(o *gwob.Obj, stride int) (float32, float32, float32, error) {
	offset := o.StrideOffsetPosition / 4
	floatsPerStride := o.StrideSize / 4
	f := offset + stride*floatsPerStride
	max := len(o.Coord)
	if f > max-3 {
		return 0, 0, 0, errors.New("overflowing vertex coord index")
	}
	return o.Coord[f], o.Coord[f+1], o.Coord[f+2], nil
}
