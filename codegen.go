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

import (
	"io"
	"fmt"
	//"bytes"
)
//import "container/list"

type GenAction uint

const (
	GA_COUNT = GenAction(1<<iota)
	GA_GENERATE
)

func revcnt_incr(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	arr,ok := i.([]int)
	if len(arr)==0 { arr = []int{0} }
	arr[len(arr)-1]++
	return arr,true
}
func revcnt_decr(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	arr,ok := i.([]int)
	if len(arr)==0 { arr = []int{0} }
	arr[len(arr)-1]--
	return arr,true
}
func revcnt_incr_first(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	arr,ok := i.([]int)
	if len(arr)==0 { arr = []int{0} }
	arr[0]++
	return arr,true
}
func revcnt_decr_first(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	arr,ok := i.([]int)
	if len(arr)==0 { arr = []int{0} }
	arr[0]--
	return arr,true
}
func revcnt_newrev(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	if !ok { return []int{0,0},true}
	arr,ok := i.([]int)
	if !ok { return []int{0,0},true}
	if len(arr)==0 { return []int{0,0},true}
	arr = append(arr,0)
	return arr,true
}
func revcnt_pullrev(i interface{},ok bool) (interface{},bool) {
	if !ok { i = []int{0} }
	arr,ok := i.([]int)
	if len(arr)!=0 { arr = arr[1:] }
	return arr,true
}

func revcnt_count_first(i interface{},ok bool) int {
	if !ok { return 0 }
	n,ok := i.([]int)
	if !ok { return 0 }
	if len(n)==0 { return 0 }
	return n[0]
}

type Generator struct{
	Dest io.Writer
	revcnt,tree ExprRefMap
	//count ExprRefMap
}
func (g *Generator) Sync(ga GenAction) {
	if ga==GA_GENERATE {
		g.tree.Iterate(func(k ExprRef,v interface{}){
			if k.SSA() { return }
			fmt.Fprintln(g.Dest,"%s = %s ;\n",k.Name,v)
		})
	}
}
func (g *Generator) Expr(e *Expr,ga GenAction) {
	switch ga {
	case GA_COUNT:
		for _,r := range e.Src {
			g.revcnt.Update(r,revcnt_incr)
		}
		if e.Flags.Has(E_HAS_DEST) {
			g.revcnt.Update(e.Dest,revcnt_newrev)
		}
	case GA_GENERATE:
		vec := make([]interface{},len(e.Src))
		for i,r := range e.Src {
			if t,ok := g.tree.Update(r,Noop); ok {
				vec[i] = t
			} else if r.SSA() {
				panic(fmt.Sprint("no such value: ",r))
			} else {
				vec[i] = r.Name
			}
			cnt := revcnt_count_first(g.revcnt.Update(r,revcnt_decr_first))
			if cnt<1 { g.tree.Delete(r) }
		}
		subtree := e.Fmt
		if e.Flags.Has(E_LITERAL) {
			if len(vec)>0 { panic("Literals must not have operands") }
		} else {
			subtree = fmt.Sprintf(e.Fmt,vec...)
		}
		if e.Flags.Has(E_HAS_DEST) {
			cnt := revcnt_count_first(g.revcnt.Update(e.Dest,revcnt_pullrev))
			if cnt==1 {
				g.tree.Update(e.Dest,Put(subtree))
			} else if e.Flags.Has(E_CHEAP) {
				g.tree.Update(e.Dest,Put(subtree))
			} else if e.Dest.SSA() {
				panic("SSA must not be used more often than one time")
			} else {
				fmt.Fprintf(g.Dest,"%s = %s ;",e.Dest.Name,subtree)
			}
		} else {
			fmt.Fprintf(g.Dest,"%s ;\n",subtree)
			
		}
	}
}

func (g *Generator) Block(b *Block,ga GenAction) {
	for e := b.Childs.Front(); e!=nil; e = e.Next() {
		if x,ok := e.Value.(*Expr); ok {
			g.Expr(x,ga)
		}
		if x,ok := e.Value.(*Block); ok {
			g.Sync(ga)
			g.Block(x,ga)
		}
	}
	g.Sync(ga)
}

