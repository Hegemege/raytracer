package models

import (
	"fmt"
	"math"
	"raytracer/utility"

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
	StartIndex int
	EndIndex   int

	SplitPlane mgl32.Vec4
}

func BuildBVH(context *RenderContext) *BVH {
	bvh := &BVH{}
	index := 0
	root := buildBVHNode(context, context.Triangles, 0, len(context.Triangles)-1, 0, &index)
	fmt.Printf("BVH has %d nodes\n", index)
	bvh.Root = root
	return bvh
}

func (bvh *BVH) Load(triangles []*Triangle) {
	bvh.Root.load(triangles)
}

func (node *BVHNode) load(triangles []*Triangle) {
	// Sorts the triangles according to the loaded BVH
	// Only for non-leaf nodes - there is no need to sort leaf nodes
	if node.LeftChild != nil {
		splitPlaneVec3 := node.SplitPlane.Vec3()
		TriangleSorter(splitPlaneVec3, triangles, node.StartIndex, node.EndIndex)

		node.LeftChild.load(triangles)
		node.RightChild.load(triangles)
	}
}

func buildBVHNode(context *RenderContext, triangles []*Triangle, startIndex int, endIndex int, depth int, index *int) *BVHNode {
	// Takes the given slice of triangles, sorts them along the major axis
	// and splits into children until leaf nodes contain the wanted number of triangles
	node := &BVHNode{
		Depth:      depth,
		Index:      *index,
		StartIndex: startIndex,
		EndIndex:   endIndex,
	}

	// Increment node index
	*index = *index + 1

	// Calculate bounds for the node
	min, max := GetTriangleBounds(triangles[startIndex:endIndex])
	node.Bounds = NewAABBMinMax(min, max)

	triCount := endIndex - startIndex + 1

	// Split if there are too many child triangles
	if context.UseBVH && triCount > context.BVHMaxLeafSize && depth < context.BVHMaxDepth {
		splitPlane := GetSplitPlaneSAH(triangles, node)
		splitPlaneVec3 := splitPlane.Vec3()

		node.SplitPlane = splitPlane

		TriangleSorter(splitPlaneVec3, triangles, startIndex, endIndex)

		splitIndex := startIndex
		splitSideSoFar := splitPlaneVec3.Dot(triangles[startIndex].Center()) > splitPlane.W()
		for i := startIndex + 1; i < endIndex; i++ { // Exclude ends
			if splitPlaneVec3.Dot(triangles[i].Center()) > splitPlane.W() != splitSideSoFar {
				splitIndex = i
				break
			}
		}

		// Fix possible errors in splitindex calc
		if splitIndex == startIndex || splitIndex == endIndex {
			splitIndex = startIndex + (endIndex-startIndex)/2
			//println("Split index is at the edge, moving to center")
		}

		node.LeftChild = buildBVHNode(context, triangles, startIndex, splitIndex-1, depth+1, index)
		node.RightChild = buildBVHNode(context, triangles, splitIndex, endIndex, depth+1, index)
	} else {
		// Add tricount to tracking
		context.BVHNodeTriangles += uint64(triCount)

		interval := uint64(len(context.Triangles) / 10.0)
		if context.BVHNodeTriangles > context.BVHProgressReported+interval {
			context.BVHProgressReported = context.BVHNodeTriangles
			progress := float32(context.BVHNodeTriangles) / float32(len(context.Triangles))
			utility.ProgressUpdate(progress, "RenderContext.BuildBVH", -1, 0)
		}
	}

	return node
}

// Returns the leaf node or nil if ray did not hit the volume
func (node *BVHNode) WalkNode(triangles []*Triangle, ray *Ray, tmin *float32, umin *float32, vmin *float32, tri **Triangle) {

	if node.LeftChild == nil {
		for _, triangle := range triangles[node.StartIndex : node.EndIndex+1] {
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
			node.LeftChild.WalkNode(triangles, ray, tmin, umin, vmin, tri)
		}

		rightHit, rightTMin, rightTMax := node.RightChild.Bounds.RayIntersect(ray)
		if rightHit && rightTMin < *tmin && rightTMax > 0 {
			node.RightChild.WalkNode(triangles, ray, tmin, umin, vmin, tri)
		}
	}

}

func GetSplitPlaneSAH(triangles []*Triangle, node *BVHNode) mgl32.Vec4 {
	// Surface area heuristic - find a split plane that minimizes the
	// surface area of the splitted AABBs

	var lowestCost float32 = math.MaxFloat32
	splitPlane := mgl32.Vec4{}

	axes := [3]mgl32.Vec3{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	triCount := node.EndIndex - node.StartIndex + 1

	for _, axis := range axes {
		// Sort the triangle slice along the given axis, compare by center
		TriangleSorter(axis, triangles, node.StartIndex, node.EndIndex)

		leftMin, leftMax := triangles[node.StartIndex].Min(), triangles[node.StartIndex].Max()
		leftAABB := MinimalAABB{Min: leftMin, Max: leftMax}
		rightMin, rightMax := triangles[node.EndIndex].Min(), triangles[node.EndIndex].Max()
		rightAABB := MinimalAABB{Min: rightMin, Max: rightMax}

		// Precompute AABB sizes, dynamic programming
		leftAABBSizes := make([]float32, triCount)
		rightAABBSizes := make([]float32, triCount)

		for i := node.StartIndex; i <= node.EndIndex; i++ {
			leftAABB.Min = utility.Vec3Min(leftAABB.Min, triangles[i].Min())
			leftAABB.Max = utility.Vec3Max(leftAABB.Max, triangles[i].Max())
			leftAABBSizes[i-node.StartIndex] = leftAABB.Area()
		}

		for i := node.EndIndex; i >= node.StartIndex; i-- {
			rightAABB.Min = utility.Vec3Min(rightAABB.Min, triangles[i].Min())
			rightAABB.Max = utility.Vec3Max(rightAABB.Max, triangles[i].Max())
			rightAABBSizes[i-node.StartIndex] = rightAABB.Area()
		}

		// Go throught every possible split along the axis to find the most optimal
		// Exclude both ends - no point in splitting 0-all
		for i := node.StartIndex + 1; i < node.EndIndex; i++ {
			leftArea := leftAABBSizes[i-node.StartIndex]
			rightArea := rightAABBSizes[i-node.StartIndex+1]

			cost := leftArea*float32(i-node.StartIndex) + rightArea*float32(node.EndIndex-i+1)
			if cost < float32(lowestCost) {
				lowestCost = cost
				var w float32
				center := triangles[i].Center()
				nextCenter := triangles[i+1].Center()
				if axis.X() > 0 {
					w = (center.X() + nextCenter.X()) / 2.0
				} else if axis.Y() > 0 {
					w = (center.Y() + nextCenter.Y()) / 2.0
				} else {
					w = (center.Z() + nextCenter.Z()) / 2.0
				}

				splitPlane = mgl32.Vec4{axis.X(), axis.Y(), axis.Z(), w}
			}
		}
	}

	return splitPlane
}
