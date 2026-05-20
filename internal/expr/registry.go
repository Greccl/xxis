package expr

var registry []*Node

func Store(node *Node) int {
	registry = append(registry, node)
	return len(registry) - 1
}

func Get(index int) *Node {
	if index < 0 || index >= len(registry) {
		return nil
	}
	return registry[index]
}
