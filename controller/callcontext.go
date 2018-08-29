package main

// Generic caller context

var DefCallContext = &CallContext{}

type CallContext struct {
	Undo bool
}

func (c *CallContext) WithUndo() *CallContext {
	cc := *c
	cc.Undo = true
	return &cc
}
