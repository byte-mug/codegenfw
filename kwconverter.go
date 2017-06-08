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

import "math/rand"
import "fmt"

// Returns true if the word is reserved.
//type WordFilter func(s string) bool

/*
Creates a Name-Changing function, that replaces reserved identifiers (as specified
trough 'wf') such as keywords with new names.
*/
func GetNameChanger(wf WordFilter) (func(string)string) {
	tl := make(map[string]string)
	am := make(map[string]bool)
	gen := rand.NewSource(12345)
	
	return func(o string) (string) {
		if t,ok := tl[o]; ok { return t }
		if a,ok := am[o]; !(a&&ok) {
			if !wf(o) { return o } /* return unchanged */
		}
		for {
			s := fmt.Sprintf("___temp_%v",gen.Int63())
			if a,ok := am[s]; a&&ok { continue } /* FAIL! Next! */
			tl[o] = s
			return s
		}
	}
}

