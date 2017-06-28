# Code Generation Library (to generate C code)

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
		blk.Childs.PushBack(codegenfw.Declare("int","a"))
		blk.Childs.PushBack(codegenfw.NewLiteral("a","7"))
		blk.Childs.PushBack(codegenfw.TouchVariable("a"))
		blk.Childs.PushBack(codegenfw.NewLiteral(1,`"Let's begin!"`))
		blk.Childs.PushBack(codegenfw.NewCall("printf(%s)",nil,1))
		blk.Childs.PushBack(codegenfw.Label("restart"))
		doif := codegenfw.CS_If_Then_Else("a")
		// if-then-else
		blk.Childs.PushBack(doif)
			doif.Childs.PushBack(codegenfw.NewLiteral(1,`"Hello!"`))
			doif.Childs.PushBack(codegenfw.NewExpr("printf(%s)",0,nil,1))
		doelse := doif.EBlock
			doelse.Childs.PushBack(codegenfw.NewLiteral(2,`"Again!"`))
			doelse.Childs.PushBack(codegenfw.NewCall("printf(%s)",nil,2))
			doelse.Childs.PushBack(codegenfw.NewLiteral(3,`1`))
			doelse.Childs.PushBack(codegenfw.NewOp("(%s-%s)","a","a",3))
			doelse.Childs.PushBack(codegenfw.GoTo("restart"))
		
		/*
		Let's do:
		a = 1
		a = a+1
		a = a+1
		a = a+1
		a = a+1
		prinft("a = %d",a);
		*/
		
		blk.Childs.PushBack(codegenfw.NewLiteral("a","1")) // a = 1
		blk.Childs.PushBack(codegenfw.NewLiteral(4,`1`)) // $4 = 1
		blk.Childs.PushBack(codegenfw.NewLiteral(5,`"a = %d"`)) // $5 = "..."
		blk.Childs.PushBack(codegenfw.NewOp("(%s+%s)","a","a",4)) //a = a+$4
		blk.Childs.PushBack(codegenfw.NewOp("(%s+%s)","a","a",4)) //a = a+$4
		blk.Childs.PushBack(codegenfw.NewOp("(%s+%s)","a","a",4)) //a = a+$4
		blk.Childs.PushBack(codegenfw.NewOp("(%s+%s)","a","a",4)) //a = a+$4
		
		blk.Childs.PushBack(codegenfw.NewCall("printf(%s,%s)",nil,5,"a")) // printf(%5,a)
	}
	fmt.Fprintln(w,"#include","<stdio.h>")
	fmt.Fprintln(w)
	fmt.Fprintln(w,"void main(){")
	{
		gen := &codegenfw.Generator{Dest:w,Indent:"\t"}
		gen.Block(blk,codegenfw.GA_COUNT)
		gen.Block(blk,codegenfw.GA_GENERATE)
	}
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

Results in...
```c
#include <stdio.h>

void main(){
	int a ;
	a = 7;
	printf("Let's begin!") ;
restart:
	if(a){
		printf("Hello!") ;
	}else{
		printf("Again!") ;
		a = (a-1) ;
		goto restart;
	}
	printf("a = %d",((((1+1)+1)+1)+1)) ;
}
```

