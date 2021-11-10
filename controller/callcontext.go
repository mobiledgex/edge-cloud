package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

// Generic caller context

type CallContext struct {
	Undo                   bool
	CRMUndo                bool
	Override               edgeproto.CRMOverride
	AutoCluster            bool
	SkipCloudletReadyCheck bool
}

func DefCallContext() *CallContext {
	return &CallContext{}
}

func (c *CallContext) WithUndo() *CallContext {
	cc := *c
	cc.Undo = true
	return &cc
}

// Normally, the CRM change is the last change in the API call,
// and if it fails, CRM will clean up after itself. Thus the
// undo function should skip any CRM changes. However, in some
// cases (like autocluster), the CRM change is not the last
// change, and we may hit other failures after the CRM change succeeds.
// In that case, we need to have the undo function apply the
// CRM changes.
func (c *CallContext) WithCRMUndo() *CallContext {
	cc := *c
	cc.CRMUndo = true
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

func (c *CallContext) Clone() *CallContext {
	clone := *c
	return &clone
}
