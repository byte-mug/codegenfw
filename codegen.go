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

const (
	curlo = "{"
	curlc = "}"
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
	Indent string
	revcnt,tree ExprRefMap
	//count ExprRefMap
	ind string
}
func (g *Generator) pushind() func() {
	o := g.ind
	g.ind = o+g.Indent
	return func(){ g.ind = o }
}
func (g *Generator) Sync(ga GenAction) {
	if ga==GA_GENERATE {
		kl := []ExprRef{}
		g.tree.Iterate(func(k ExprRef,v interface{}){
			if k.SSA() { return }
			fmt.Fprintf(g.Dest,"%s%s = %s ;\n",g.ind,k.Name,v)
			kl = append(kl,k)
		})
		for _,k := range kl { g.tree.Delete(k) }
	}
}

func (g *Generator) vec(src []ExprRef,ga GenAction) []interface{} {
	switch ga {
	case GA_COUNT:
		for _,r := range src {
			g.revcnt.Update(r,revcnt_incr)
		}
	case GA_GENERATE:
		vec := make([]interface{},len(src))
		for i,r := range src {
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
		return vec
	}
	return nil
}
func (g *Generator) Expr(e *Expr,ga GenAction) {
	vec := g.vec(e.Src,ga)
	switch ga {
	case GA_COUNT:
		if e.Flags.Has(E_HAS_DEST) {
			g.revcnt.Update(e.Dest,revcnt_newrev)
		}
	case GA_GENERATE:
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
				fmt.Fprintf(g.Dest,"%s%s = %s ;",g.ind,e.Dest.Name,subtree)
			}
		} else {
			fmt.Fprintf(g.Dest,"%s%s ;\n",g.ind,subtree)
			
		}
	}
}

func (g *Generator) Block(b *Block,ga GenAction) {
	defer g.pushind()()
	for e := b.Childs.Front(); e!=nil; e = e.Next() {
		if x,ok := e.Value.(*Expr); ok {
			g.Expr(x,ga)
		}
		if x,ok := e.Value.(*Block); ok {
			g.Sync(ga)
			g.Block(x,ga)
		}
		
		if x,ok := e.Value.(*ControlStruct); ok {
			g.Sync(ga)
			vec := g.vec(x.Src,ga)
			switch ga{
			case GA_GENERATE:
				fmt.Fprintf(g.Dest,"%s",g.ind)
				fmt.Fprintf(g.Dest,x.Fmt,vec...)
				fmt.Fprintln(g.Dest,curlo)
			}
			g.Block(&x.Block,ga)
			vec = g.vec(x.Src2,ga)
			switch ga{
			case GA_GENERATE:
				fmt.Fprintf(g.Dest,"%s%s",g.ind,curlc)
				fmt.Fprintf(g.Dest,x.Fmt2,vec...)
				if x.EBlock!=nil {
					fmt.Fprintln(g.Dest,curlo)
				}else{
					fmt.Fprintln(g.Dest)
				}
			}
			if x.EBlock!=nil {
				g.Block(x.EBlock,ga)
				switch ga{
				case GA_GENERATE:
					fmt.Fprintf(g.Dest,"%s%s",g.ind,curlc)
					fmt.Fprintln(g.Dest)
				}
			}
		}
		
	}
	g.Sync(ga)
}

