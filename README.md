# Code Generation Framework (to generate C code)

Not usable yet......


```go
package main

import "os"
import "io"
import "fmt"
import "github.com/byte-mug/codegenfw"

func out(w io.Writer) {
	blk := new(codegenfw.Block)
	blk.Childs.Init()
	{
		blk.Childs.PushBack(codegenfw.NewLiteral("a",`1`))
		doif := codegenfw.CS_If_Then_Else("a")
		// if-then-else
		blk.Childs.PushBack(doif)
			doif.Childs.PushBack(codegenfw.NewLiteral(1,`"Hello!"`))
			doif.Childs.PushBack(codegenfw.NewExpr("printf(%s)",0,nil,1))
		doelse := doif.EBlock
			doelse.Childs.PushBack(codegenfw.NewLiteral(2,`"Hello!"`))
			doelse.Childs.PushBack(codegenfw.NewExpr("printf(%s)",0,nil,2))
	}
	fmt.Fprintln(w,"#include","<stdio.h>")
	fmt.Fprintln(w)
	fmt.Fprintln(w,"void main(){")
	fmt.Fprintln(w,"int a;")
	{
		gen := &codegenfw.Generator{Dest:w,Indent:"\t"}
		gen.Block(blk,codegenfw.GA_COUNT)
		gen.Block(blk,codegenfw.GA_GENERATE)
	}
	//fmt.Fprintln(w,`printf("Hello!\n");`)
	fmt.Fprintln(w,"}")
	fmt.Fprintln(w)
	fmt.Fprintln(w)
}

func main() {
	f,e := os.Create("program.c")
	if e!=nil { fmt.Println(e); return }
	defer f.Close()
	out(f)
}
```