package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

// Generic caller context

type CallContext struct {
	Undo        bool
	Override    edgeproto.CRMOverride
	AutoCluster bool
}

func DefCallContext() *CallContext {
	return &CallContext{}
}

func (c *CallContext) WithUndo() *CallContext {
	cc := *c
	cc.Undo = true
	return &cc
}

func (c *CallContext) WithAutoCluster() *CallContext {
	cc := *c
	cc.AutoCluster = true
	return &cc
}

// SetOverride takes the override specified from the user,
// and removes it from the input object.
// Because there may be multiple calls to this function,
// we only modify the override if it's non-default.
// Override is only meant as a switch to the current operation,
// not as a persistent state on the object.
func (c *CallContext) SetOverride(o *edgeproto.CRMOverride) {
	if *o == edgeproto.CRMOverride_NO_OVERRIDE {
		return
	}
	c.Override = *o
	*o = edgeproto.CRMOverride_NO_OVERRIDE
}
