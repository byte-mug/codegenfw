/*
MIT License

Copyright (c) 2017 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package codegenfw

import "container/list"

type EFlags uint

func (e EFlags) Has(i EFlags) bool { return (e&i)==i }

const (
	E_HAS_DEST = EFlags(1<<iota)
	E_CHEAP
	E_LITERAL
)

type ExprRef struct{
	Name string
	Num uint
}
func (e ExprRef) SSA() bool { return len(e.Name)==0 }
func NewExprRef(i interface{}) ExprRef {
	switch v := i.(type) {
	case string: return ExprRef{v,0}
	case uint: return ExprRef{"",v}
	case int: return ExprRef{"",uint(v)}
	}
	panic("illegal value")
}


type ExprRefMap struct{
	names map[string]interface{}
	nums map[uint]interface{}
}
func (e *ExprRefMap) Update(k ExprRef, f func(interface{},bool) (interface{},bool)) (interface{},bool) {
	if len(k.Name)==0 {
		if e.nums==nil { e.nums = make(map[uint]interface{}) }
		i,ok := e.nums[k.Num]
		i,ok = f(i,ok)
		if ok { e.nums[k.Num] = i }
		return i,ok
	}
	if e.names==nil { e.names = make(map[string]interface{}) }
	i,ok := e.names[k.Name]
	i,ok = f(i,ok)
	if ok { e.names[k.Name] = i }
	return i,ok
}
func (e *ExprRefMap) Delete(k ExprRef) {
	if len(k.Name)==0 {
		if e.nums==nil { return }
		delete(e.nums,k.Num)
	}
	if e.names==nil { return }
	delete(e.names,k.Name)
}
func (e *ExprRefMap) Iterate(f func(k ExprRef,v interface{})) {
	if e.nums!=nil {
		for n,v := range e.nums { f(ExprRef{"",n},v) }
	}
	if e.names!=nil {
		for n,v := range e.names { f(ExprRef{n,0},v) }
	}
}


func Noop(i interface{},ok bool) (interface{},bool) { return i,ok }

func Incr(i interface{},ok bool) (interface{},bool) {
	n := 0
	if ok {
		if m,ok := i.(int); ok { n = m }
	}
	n++
	if n<=0 { return nil,false }
	return n,true
}
func Decr(i interface{},ok bool) (interface{},bool) {
	n := 0
	if ok {
		if m,ok := i.(int); ok { n = m }
	}
	if n<=0 { return nil,false }
	n--
	return n,true
}

func Put(i interface{}) (func(i interface{},ok bool) (interface{},bool)) {
	return func(interface{},bool) (interface{},bool) { return i,true }
}


type Expr struct{
	Dest ExprRef
	Src []ExprRef
	Fmt string
	Flags EFlags
}
func NewExpr(fmt string,f EFlags,dst interface{}, src ...interface{}) (result *Expr) {
	result = new(Expr)
	result.Fmt = fmt
	result.Flags = f
	if dst!=nil {
		result.Flags |= E_HAS_DEST
		result.Dest = NewExprRef(dst)
	} else { result.Flags &= ^E_HAS_DEST }
	if l := len(src); l>0 {
		result.Src = make([]ExprRef,l)
		for i,srci := range src { result.Src[i] = NewExprRef(srci) }
	}
	return
}
func NewLiteral(dst interface{},val string) *Expr {
	return &Expr{NewExprRef(dst),nil,val,E_HAS_DEST|E_LITERAL|E_CHEAP}
}

type Block struct{
	Childs list.List
}

