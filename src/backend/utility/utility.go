package utility

import (
	"encoding/json"
	"math"
	"math/rand"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
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
