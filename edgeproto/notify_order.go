package edgeproto

// NotifyOrder codifies the order in which objects must be sent through
// notify, based on their inter-dependencies.
// Objects that depend on other objects must be sent after the objects they
// depend on. This guarantees that if the downstream service wants to look
// up the dependency, it will have already been received.
// Lower order numbers represent sending before higher order numbers.
// Sorting by Less functions means dependencies will be sorted first.

// NotifyOrderObj is the object which will have order constraints
type NotifyOrderObj interface {
	GetTypeString() string
}

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
	// Cloudlet dependencies
	n.addObjectDep(&CloudletCache{}, &FlavorCache{})
	n.addObjectDep(&CloudletCache{}, &VMPoolCache{})
	n.addObjectDep(&CloudletCache{}, &GPUDriverCache{})
	// ClusterInst dependencies
	n.addObjectDep(&ClusterInstCache{}, &FlavorCache{})
	n.addObjectDep(&ClusterInstCache{}, &CloudletCache{})
	n.addObjectDep(&ClusterInstCache{}, &AutoScalePolicyCache{})
	n.addObjectDep(&ClusterInstCache{}, &TrustPolicyCache{})
	n.addObjectDep(&ClusterInstCache{}, &NetworkCache{})
	// App dependencies
	n.addObjectDep(&AppCache{}, &FlavorCache{})
	n.addObjectDep(&AppCache{}, &AutoProvPolicyCache{})
	// AppInst dependencies
	n.addObjectDep(&AppInstCache{}, &FlavorCache{})
	n.addObjectDep(&AppInstCache{}, &AppCache{})
	n.addObjectDep(&AppInstCache{}, &ClusterInstCache{})
	n.addObjectDep(&AppInstCache{}, &TrustPolicyCache{})
	// AppInstRefs dependencies
	n.addObjectDep(&AppInstRefsCache{}, &AppCache{})
	n.addObjectDep(&AppInstRefsCache{}, &AppInstCache{})
	n.addObjectDep(&AppInstRefsCache{}, &CloudletCache{})
	// CloudletRefs dependencies
	n.addObjectDep(&CloudletRefsCache{}, &ClusterInstCache{})
	n.addObjectDep(&CloudletRefsCache{}, &AppInstCache{})
	// ClusterRefs dependencies
	n.addObjectDep(&ClusterRefsCache{}, &AppInstCache{})
	// TrustPolicyException dependencies
	n.addObjectDep(&TrustPolicyExceptionCache{}, &AppCache{})
	n.addObjectDep(&TrustPolicyExceptionCache{}, &CloudletPoolCache{})
	// CloudletPool dependencies
	n.addObjectDep(&CloudletPoolCache{}, &CloudletCache{})
	return n
}

func (s *NotifyOrder) getNode(obj NotifyOrderObj) *NotifyOrderNode {
	node, ok := s.objs[obj.GetTypeString()]
	if !ok {
		node = &NotifyOrderNode{
			typeName: obj.GetTypeString(),
		}
		s.objs[obj.GetTypeString()] = node
	}
	return node
}

func (s *NotifyOrder) addObjectDep(obj, dependsOn NotifyOrderObj) {
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

// Get types and the types they depend upon
func (s *NotifyOrder) GetDeps() map[string][]string {
	deps := make(map[string][]string)
	for _, obj := range s.objs {
		objDeps, _ := deps[obj.typeName]
		for _, node := range obj.dependsOn {
			objDeps = append(objDeps, node.typeName)
		}
		deps[obj.typeName] = objDeps
	}
	return deps
}
