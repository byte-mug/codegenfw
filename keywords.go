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

// Returns true if the word is reserved.
type WordFilter func(s string) bool

func hasPrefix(a,p string) bool {
	la,lp := len(a),len(p)
	if la<lp { return false }
	return a[:lp]==p
}

func WF_ANSI_C(s string) bool {
	switch s {
	case	"auto","break","case","char","const","continue","default","do","double","else",
		"enum","extern","float","for","goto","if","int","long","register","return","short","signed","sizeof","static",
		"struct","switch","typedef","union","unsigned","void","volatile","while": return true
	}
	return false
}

// WF_ANSI_C + asm typeof inline 
func WF_ModernC(s string) bool {
	switch s {
	case	"auto","break","case","char","const","continue","default","do","double","else",
		"enum","extern","float","for","goto","if","int","long","register","return","short","signed","sizeof","static",
		"struct","switch","typedef","union","unsigned","void","volatile","while":
			return true
	case	"asm","typeof","inline":
			return true
	}
	return false
}

// This is a very Picky keyword filter, trying to cut off all GCC keywords.
func WF_GccC(s string) bool {
	switch s {
	case	"auto","break","case","char","const","continue","default","do","double","else",
		"enum","extern","float","for","goto","if","int","long","register","return","short","signed","sizeof","static",
		"struct","switch","typedef","union","unsigned","void","volatile","while":
			return true
	case	"asm","typeof","inline":
			return true
	case	"__attribute__","__complex__", "__declspec","__ea","__extension__","__far","__imag__","__real__","__memx",
		"__thread","__func__","__asm__":
			return true
	case	"__FUNCTION__","__PRETTY_FUNCTION__","__STDC_HOSTED__":
			return true
	case	"__FILE__","__LINE__": return true
	}
	
	ls := len(s)
	
	if ls>=2 {
		if s[0]=='_' {
			m := s[1]
			if (m>='A') && (m<='Z') { /* exclude "_" + UpperCaseChar */ return true }
			switch m {
				// restrict _exit, _xabort, _xbegin, _xend, _xtest
				case 'e','x': return true
			}
		}
	}
	
	if ls>=3 {
		if s[:2]=="__" {
			m := s[2]
			//if (m>='A') && (m<='Z') { /* exclude "__" + UpperCaseChar */ return true }
			if (m!='_') && ((m<'0')||(m>'9')) {
				n := s[2:] // get rid of the "__"-prefix!
				if hasPrefix(n,"atomic") { return true }
				if hasPrefix(n,"builtin") { return true }
				if hasPrefix(n,"sync") { return true }
				if hasPrefix(n,"flash") { return true }
				if hasPrefix(n,"float") { return true }
				if hasPrefix(n,"fp") { return true }
				if hasPrefix(n,"int") { return true }
			}
		}
	}
	return false
}


