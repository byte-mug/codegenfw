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
	GA_COUNT = GenAction(iota)
	GA_GENERATE
)

const (
	curlo = "{"
	curlc = "}"
)

type GenFlags uint
func (a GenFlags) Has(b GenFlags) bool { return (a&b)==b }
const (
	GF_ENFORCE_STORE = GenFlags(1<<iota)
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
func eflags_cast(i interface{},ok bool) EFlags {
	j,ok2 := i.(EFlags)
	if ok&&ok2 { return j }
	return 0
}

func bool_cast(i interface{},ok bool) bool {
	j,ok2 := i.(bool)
	return ok && ok2 && j
}

func (e EFlags) set(i interface{},ok bool) (interface{},bool) {
	return (e|eflags_cast(i,ok)),true
}
func (e EFlags) clear(i interface{},ok bool) (interface{},bool) {
	return (eflags_cast(i,ok)&^e),true
}

type Generator struct{
	Dest io.Writer
	Indent string
	Flags GenFlags
	revcnt,tree,eflags,volm ExprRefMap
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

func (g *Generator) vec(src []ExprRef,ga GenAction) ([]interface{},EFlags,EFlags) {
	switch ga {
	case GA_COUNT:
		for _,r := range src {
			g.revcnt.Update(r,revcnt_incr)
		}
	case GA_GENERATE:
		vec := make([]interface{},len(src))
		aflags := ^EFlags(0)
		oflags := EFlags(0)
		for i,r := range src {
			if t,ok := g.tree.Update(r,Noop); ok {
				vec[i] = t
				temp1 := eflags_cast(g.eflags.Update(r,Noop))
				oflags |= temp1
				aflags &= temp1
			} else if r.SSA() {
				panic(fmt.Sprint("no such value: ",r))
			} else {
				vec[i] = r.Name
			}
			cnt := revcnt_count_first(g.revcnt.Update(r,revcnt_decr_first))
			if cnt<1 {
				g.tree.Delete(r)
				g.eflags.Delete(r)
			}
		}
		return vec,oflags,aflags
	}
	return nil,0,0
}
func (g *Generator) Expr(e *Expr,ga GenAction) {
	vec,oflags,aflags := g.vec(e.Src,ga)
	oflags |= e.Flags
	aflags &= e.Flags
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
			nflags := efjoin(oflags,aflags)
			
			// Whether Store is required or not.
			require_store := g.Flags.Has(GF_ENFORCE_STORE)||
				// Time-Critical instructions must be evaluated immediately.
				nflags.Has(E_TIME_CRITICAL)||
				
				// if the destination variable is volatile, store is required
				bool_cast(g.volm.Update(e.Dest,Noop))
			
			// wether the code must be evaluated or not.
			must_evaluate := nflags.Has(E_NO_OMIT)||
				nflags.Has(E_TIME_CRITICAL)
			
			if require_store && !e.Dest.SSA() {
				fmt.Fprintf(g.Dest,"%s%s = %s ;\n",g.ind,e.Dest.Name,subtree)
			} else if cnt==1 {
				g.tree.Update(e.Dest,Put(subtree))
				g.eflags.Update(e.Dest,Put(nflags))
			} else if nflags.Has(E_CHEAP) {
				g.tree.Update(e.Dest,Put(subtree))
				g.eflags.Update(e.Dest,Put(nflags))
			} else if e.Dest.SSA() {
				if cnt==0 { // oops... the code doesn't consume the value.
					
					// In this case, the code must be evaluated.
					if must_evaluate {
						fmt.Fprintf(g.Dest,"%s%s ;\n",g.ind,subtree)
					}
				}
				panic("When SSA, value count must be 1")
			} else {
				fmt.Fprintf(g.Dest,"%s%s = %s ;\n",g.ind,e.Dest.Name,subtree)
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
			if ga == GA_GENERATE {
				fmt.Fprintf(g.Dest,"%s%s\n",g.ind,curlo)
			}
			g.Block(x,ga)
			if ga == GA_GENERATE {
				fmt.Fprintf(g.Dest,"%s%s\n",g.ind,curlc)
			}
		}
		
		if x,ok := e.Value.(*ControlStruct); ok {
			g.Sync(ga)
			vec,_,_ := g.vec(x.Src,ga)
			if ga == GA_GENERATE {
				fmt.Fprintf(g.Dest,"%s",g.ind)
				fmt.Fprintf(g.Dest,x.Fmt,vec...)
				fmt.Fprintln(g.Dest,curlo)
			}
			g.Block(&x.Block,ga)
			vec,_,_ = g.vec(x.Src2,ga)
			if ga == GA_GENERATE {
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
				if ga == GA_GENERATE {
					fmt.Fprintf(g.Dest,"%s%s",g.ind,curlc)
					fmt.Fprintln(g.Dest)
				}
			}
		}
		if x,ok := e.Value.(Label); ok {
			g.Sync(ga)
			if ga == GA_GENERATE {
				fmt.Fprintf(g.Dest,"%s:\n",string(x))
			}
		}
		if x,ok := e.Value.(GoTo); ok {
			g.Sync(ga)
			if ga == GA_GENERATE {
				fmt.Fprintf(g.Dest,"%sgoto %s;\n",g.ind,string(x))
			}
		}
		if x,ok := e.Value.(EnforceStore); ok {
			if (ga == GA_GENERATE) && len(x)>0 {
				er := ExprRef{string(x),0}
				s,ok := g.tree.Update(er,Noop)
				if ok {
					fmt.Fprintf(g.Dest,"%s%s = %s;\n",g.ind,er.Name,s)
					g.tree.Delete(er)
				}
			}
		}
		if x,ok := e.Value.(TouchVariable); ok {
			if len(x)>0 {
				er := ExprRef{string(x),0}
				switch ga {
				case GA_COUNT:
					g.revcnt.Update(er,revcnt_incr)
				case GA_GENERATE:
					g.revcnt.Update(er,revcnt_decr_first)
					s,ok := g.tree.Update(er,Noop)
					if ok {
						fmt.Fprintf(g.Dest,"%s%s = %s;\n",g.ind,er.Name,s)
						g.tree.Delete(er)
					}
				}
			}
		}
		if x,ok := e.Value.(*Declaration); ok {
			if (ga == GA_GENERATE) && len(x.Names)>0 {
				fmt.Fprintf(g.Dest,"%s%s %s",g.ind,x.DataType,x.Names[0])
				for _,n := range x.Names[1:] { fmt.Fprintf(g.Dest," ,%s",n) }
				fmt.Fprintln(g.Dest," ;")
			}
		}
		if x,ok := e.Value.(SetVolatile); ok {
			if (ga == GA_GENERATE) && len(x.Variable)>0 {
				er := ExprRef{x.Variable,0}
				g.volm.Update(er,Put(x.Volatile))
				
				/*
				If the variable is now volatile, enforce store.
				*/
				if x.Volatile {
					s,ok := g.tree.Update(er,Noop)
					if ok {
						fmt.Fprintf(g.Dest,"%s%s = %s;\n",g.ind,er.Name,s)
						g.tree.Delete(er)
						g.eflags.Delete(er)
					}
				}
			}
		}
	}
	g.Sync(ga)
}

