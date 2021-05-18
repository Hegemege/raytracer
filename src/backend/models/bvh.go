package models

import (
	"fmt"
	"math"
	"raytracer/utility"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
)

// Bounding volume hierarchy for accelerated
// ray hit detection
type BVH struct {
	Root *BVHNode
}

type BVHNode struct {
	Depth      int
	Index      int
	LeftChild  *BVHNode
	RightChild *BVHNode
	Bounds     *AABB

	// Leaf nodes will contain references to all triangles within the node
	Triangles []*Triangle
}

func BuildBVH(context *RenderContext, triangles []*Triangle) *BVH {
	bvh := &BVH{}
	index := 0
	root := buildBVHNode(context, triangles, 0, &index)
	fmt.Printf("BVH has %d nodes\n", index)
	bvh.Root = root
	return bvh
}

func buildBVHNode(context *RenderContext, triangles []*Triangle, depth int, index *int) *BVHNode {
	// Takes the given slice of triangles, sorts them along the major axis
	// and splits into children until leaf nodes contain the wanted number of triangles
	node := &BVHNode{
		Depth: depth,
		Index: *index,
	}

	// Increment node index
	*index = *index + 1

	// Calculate bounds for the node
	min, max := GetTriangleBounds(triangles)
	node.Bounds = NewAABBMinMax(min, max)
	node.Triangles = triangles[:]

	triCount := len(triangles)

	// Split if there are too many child triangles
	if triCount > context.Settings.BVHMaxLeafSize && context.Settings.UseBVH {
		splitPlane := GetSplitPlaneSAH(node)
		splitPlaneVec3 := splitPlane.Vec3()

		sort.SliceStable(node.Triangles, func(i, j int) bool {
			return splitPlaneVec3.Dot(node.Triangles[i].Center) < splitPlaneVec3.Dot(node.Triangles[j].Center)
		})

		splitIndex := 0
		splitSideSoFar := splitPlaneVec3.Dot(node.Triangles[0].Center) > splitPlane.W()
		for i := 1; i < triCount; i++ {
			if splitPlaneVec3.Dot(node.Triangles[i].Center) > splitPlane.W() != splitSideSoFar {
				splitIndex = i
				break
			}
		}

		// Fix possible errors in splitindex calc
		if splitIndex == 0 || splitIndex == triCount-1 {
			splitIndex = triCount / 2
			//println("Split index is at the edge, moving to center")
		}

		leftTriangles := node.Triangles[:splitIndex]
		node.LeftChild = buildBVHNode(context, leftTriangles, depth+1, index)

		rightTriangles := node.Triangles[splitIndex:]
		node.RightChild = buildBVHNode(context, rightTriangles, depth+1, index)
	}

	return node
}

// Returns the leaf node or nil if ray did not hit the volume
func (node *BVHNode) WalkNode(ray *Ray, tmin *float32, umin *float32, vmin *float32, tri **Triangle) {

	if node.LeftChild == nil {
		for _, triangle := range node.Triangles {
			if triangle.Normal.Dot(ray.Direction) > 0 {
				continue
			}
			t, u, v := triangle.RayIntersect(ray)
			if t > 0 && t < *tmin {
				*tmin = t
				*umin = u
				*vmin = v
				*tri = triangle
			}
		}
	} else {
		leftHit, leftTMin, leftTMax := node.LeftChild.Bounds.RayIntersect(ray)
		if leftHit && leftTMin < *tmin && leftTMax > 0 {
			node.LeftChild.WalkNode(ray, tmin, umin, vmin, tri)
		}

		rightHit, rightTMin, rightTMax := node.RightChild.Bounds.RayIntersect(ray)
		if rightHit && rightTMin < *tmin && rightTMax > 0 {
			node.RightChild.WalkNode(ray, tmin, umin, vmin, tri)
		}
	}

}

func GetSplitPlaneSAH(node *BVHNode) mgl32.Vec4 {
	// Surface area heuristic - find a split plane that minimizes the
	// surface area of the splitted AABBs

	var lowestCost float32 = math.MaxFloat32
	splitPlane := mgl32.Vec4{}

	axes := [3]mgl32.Vec3{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	triCount := len(node.Triangles)

	for _, axis := range axes {
		// Sort the triangle slice along the given axis, compare by center
		sort.SliceStable(node.Triangles, func(i, j int) bool {
			return axis.Dot(node.Triangles[i].Center) < axis.Dot(node.Triangles[j].Center)
		})

		leftMin, leftMax := node.Triangles[0].Min, node.Triangles[0].Max
		leftAABB := MinimalAABB{Min: leftMin, Max: leftMax}
		rightMin, rightMax := node.Triangles[triCount-1].Min, node.Triangles[triCount-1].Max
		rightAABB := MinimalAABB{Min: rightMin, Max: rightMax}

		// Precompute AABB sizes, dynamic programming
		leftAABBSizes := make([]float32, triCount)
		rightAABBSizes := make([]float32, triCount)

		for i := 0; i < triCount; i++ {
			leftAABB.Min = utility.Vec3Min(leftAABB.Min, node.Triangles[i].Min)
			leftAABB.Max = utility.Vec3Max(leftAABB.Max, node.Triangles[i].Max)
			leftAABBSizes[i] = leftAABB.Area()
		}

		for i := triCount - 1; i >= 0; i-- {
			rightAABB.Min = utility.Vec3Min(rightAABB.Min, node.Triangles[i].Min)
			rightAABB.Max = utility.Vec3Max(rightAABB.Max, node.Triangles[i].Max)
			rightAABBSizes[i] = rightAABB.Area()
		}

		// Go throught every possible split along the axis to find the most optimal
		// Exclude both ends - no point in splitting 0-all
		for i := 1; i < triCount-1; i++ {
			leftArea := leftAABBSizes[i]
			rightArea := rightAABBSizes[i+1]

			cost := leftArea*float32(i) + rightArea*float32(triCount-i)
			if cost < float32(lowestCost) {
				lowestCost = cost
				var w float32
				if axis.X() > 0 {
					w = (node.Triangles[i].Center.X() + node.Triangles[i+1].Center.X()) / 2.0
				} else if axis.Y() > 0 {
					w = (node.Triangles[i].Center.Y() + node.Triangles[i+1].Center.Y()) / 2.0
				} else {
					w = (node.Triangles[i].Center.Z() + node.Triangles[i+1].Center.Z()) / 2.0
				}

				splitPlane = mgl32.Vec4{axis.X(), axis.Y(), axis.Z(), w}
			}
		}
	}

	return splitPlane
}
