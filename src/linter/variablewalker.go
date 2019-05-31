package linter

/*

When does the context state switch from its current value?

- The left hand side of any assignment type operation starts a write.
- A list() expression starts a write for each of its item values.
- A subexpression for a variable name (e.g., $$x) is always a read.
- Array and object writes subsume all their modifications.



*/

import (
	"fmt"
	"sort"

	"github.com/VKCOM/noverify/src/meta"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

type VariableContext struct {
	parent  walker.Walkable
	key     string
	isWrite bool
}

type VariableContextStack []VariableContext

func (s *VariableContextStack) Peek() VariableContext {
	if len(*s) == 0 {
		return VariableContext{}
	}
	return (*s)[len(*s)-1]
}

func (s *VariableContextStack) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *VariableContextStack) Push(parent walker.Walkable, key string, isWrite bool) {
	*s = append(*s, VariableContext{parent, key, isWrite})
}

type VariableWalker struct {
	reads    map[string]struct{}
	writes   map[string]struct{}
	varStack VariableContextStack
}

func (vw *VariableWalker) addRead(v string) {
	if vw.reads == nil {
		vw.reads = make(map[string]struct{})
	}
	vw.reads[v] = struct{}{}
}

func (vw *VariableWalker) addWrite(v string) {
	if vw.writes == nil {
		vw.writes = make(map[string]struct{})
	}
	vw.writes[v] = struct{}{}
}

func (vw *VariableWalker) Reads() []string {
	vs := make([]string, 0, len(vw.reads))
	for v, _ := range vw.reads {
		vs = append(vs, v)
	}
	sort.Strings(vs)
	return vs
}

func (vw *VariableWalker) Writes() []string {
	vs := make([]string, 0, len(vw.writes))
	for v, _ := range vw.writes {
		vs = append(vs, v)
	}
	sort.Strings(vs)
	return vs
}

func (vw *VariableWalker) EnterNode(w walker.Walkable) bool {
	if !meta.IsIndexingComplete() {
		return false
	}

	var id *node.Identifier

	switch n := w.(type) {
	case *node.Identifier:
		id = n
	case *stmt.Property, *node.Parameter:
		return false
	default:
		return true
	}

	ctx := vw.varStack.Peek()
	if _, ok := ctx.parent.(*expr.Variable); ok {
		if ctx.isWrite {
			vw.addWrite(id.Value)
		} else {
			vw.addRead(id.Value)
		}
	}
	return true
}

func (vw *VariableWalker) LeaveNode(w walker.Walkable) {}

func (vw *VariableWalker) EnterChildNode(key string, w walker.Walkable) {
	switch n := w.(type) {
	case *assign.Assign:
		if key == "Variable" {
			vw.varStack.Push(w, key, true)
		} else {
			vw.varStack.Push(w, key, false)
		}
	case *expr.PropertyFetch:
		if key == "Property" {
			vw.varStack.Push(w, key, false)
		}
	case *expr.ArrayDimFetch:
		if key == "Dim" {
			vw.varStack.Push(w, key, false)
		}
	case *expr.Variable:
		if key != "VarName" {
			panic(fmt.Sprintf("Unexpected child key '%s' for expr.Variable", key))
		}
		if _, ok := n.VarName.(*node.Identifier); ok {
			isWrite := vw.varStack.Peek().isWrite
			vw.varStack.Push(w, key, isWrite)
		} else {
			vw.varStack.Push(w, key, false)
		}
	}
}

func (vw *VariableWalker) LeaveChildNode(key string, w walker.Walkable) {
	vctx := vw.varStack.Peek()
	if vctx.parent == w && vctx.key == key {
		vw.varStack.Pop()
	}
}

func (vw *VariableWalker) EnterChildList(key string, w walker.Walkable) {}

func (vw *VariableWalker) LeaveChildList(key string, w walker.Walkable) {}
