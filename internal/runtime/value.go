package runtime

import (
   "sync"
)

type Value interface {
	Expand() string
	Reset()
	Append(string)
	Method(string)
}

type SyncValue struct {
   val Value
   mu sync.Mutex
}

//
// String
//

func NewStringValue(value string) *StringValue {
	v := StringValue(value)
	return &v
}

type StringValue string

func (v *StringValue) Expand() string {
	return string(*v)
}

func (v *StringValue) Reset() {
	*v = ""
}

func (v *StringValue) Append(s string) {
	*v = StringValue(string(*v) + s)
}

func (v *StringValue) Method(string) {
}

//
// Module
//

type ModuleValue struct {
	Module *Module
}

func (v ModuleValue) Expand() string {
	if v.Module == nil {
		return ""
	}
	return v.Module.Path
}

func (v ModuleValue) Get(name string) string {
	if v.Module == nil {
		return ""
	}
	if item, ok := v.Module.Locals[name]; ok {
		return item.Expand()
	}
	/*
	if item, ok := v.Module.Shared.Get(name); ok {
		return item.Expand()
	}
	*/
	return ""
}

func (v ModuleValue) Set(name, value string) {
	if v.Module == nil {
		return
	}
	v.Module.Locals[name] = NewStringValue(value)
}

func (v ModuleValue) Reset() {
}

func (v ModuleValue) Append(string) {
}

func (v ModuleValue) Method(string) {
}
