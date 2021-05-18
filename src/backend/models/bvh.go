package models

// Bounding volume hierarchy for accelerated
// ray hit detection
type BVH struct {
	Root *BVHNode
}

type BVHNode struct {
	Depth    int
	Index    int
	Children []*BVHNode
	Bounds   *AABB

	// Leaf nodes will contain references to all triangles within the node
	Triangles []*Triangle
}

func BuildBVH(triangles []*Triangle) *BVH {
	bvh := &BVH{}
	index := 0
	root := buildBVHNode(triangles, 0, &index)
	println("new index", index)
	bvh.Root = root
	return bvh
}

func buildBVHNode(triangles []*Triangle, depth int, index *int) *BVHNode {
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
	bounds := NewAABBMinMax(min, max)
	node.Bounds = bounds

	// TODO: Split
	node.Triangles = triangles
	node.Children = nil

	return node
}
