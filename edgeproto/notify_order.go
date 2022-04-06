package edgeproto

// NotifyOrder codifies the order in which objects must be sent through
// notify, based on their inter-dependencies.
// Objects that depend on other objects must be sent after the objects they
// depend on. This guarantees that if the downstream service wants to look
// up the dependency, it will have already been received.
// Lower order numbers represent sending before higher order numbers.
// Sorting by Less functions means dependencies will be sorted first.

type NotifyOrder struct {
	// key is cache type (cache.GetTypeString())
	objs map[string]*NotifyOrderNode
}

type NotifyOrderNode struct {
	typeName     string
	order        int
	dependsOn    []*NotifyOrderNode
	dependedOnBy []*NotifyOrderNode
}

func NewNotifyOrder() *NotifyOrder {
	n := &NotifyOrder{}
	n.objs = make(map[string]*NotifyOrderNode)
	for obj, refs := range GetReferencesMap() {
		for _, ref := range refs {
			n.addObjectDep(obj, ref)
		}
	}
	return n
}

func (s *NotifyOrder) getNode(obj string) *NotifyOrderNode {
	node, ok := s.objs[obj]
	if !ok {
		node = &NotifyOrderNode{
			typeName: obj,
		}
		s.objs[obj] = node
	}
	return node
}

func (s *NotifyOrder) addObjectDep(obj, dependsOn string) {
	objNode := s.getNode(obj)
	depNode := s.getNode(dependsOn)

	objNode.dependsOn = append(objNode.dependsOn, depNode)
	depNode.dependedOnBy = append(depNode.dependedOnBy, objNode)

	if depNode.order < objNode.order {
		// no change
		return
	}
	objNode.order = depNode.order + 1
	objNode.orderUpdated()
}

func (s *NotifyOrderNode) orderUpdated() {
	// propagate change to order to nodes depending on this node
	for i := range s.dependedOnBy {
		if s.order < s.dependedOnBy[i].order {
			// no change needed, already higher order
			continue
		}
		s.dependedOnBy[i].order = s.order + 1
		s.dependedOnBy[i].orderUpdated()
	}
}

func (s *NotifyOrder) Less(typeString1, typeString2 string) bool {
	var order1, order2 int
	if node, ok := s.objs[typeString1]; ok {
		order1 = node.order
	}
	if node, ok := s.objs[typeString2]; ok {
		order2 = node.order
	}
	return order1 < order2
}
