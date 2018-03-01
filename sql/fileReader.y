// Copyright © 2017 Mike Hudgins <mchudgins@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// based off of Appendix A from http://dinosaur.compilertools.net/yacc/

%{

package sql

import (
    "fmt"
    "io/ioutil"
    "strconv"
    "strings"
    "unicode"
)

%}

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union{
records []Record
record Record
val int
str string
}

// any non-terminal which returns a value needs a type, which is
// really a field name in the above union struct
%type <records> recordList
%type <record> record
%type <str> project
%type <str> page
%type <str> hits
%type <str> size

// same for terminals
%token <str> STRING NL NUMBER

%%

file : recordList { records = $1 }
    ;

recordList : /* empty */ { $$ = make([]Record, 0, 2000000) }
    | recordList record {
                        $$ = append( $1, $2 )
                        }
    ;

record : project page hits size NL  {
                                    $$.Project = $1
                                    $$.Page = $2
                                    $$.Hits, _ = strconv.Atoi($3)
                                    $$.Size, _ = strconv.Atoi( $4 )
                                    }
    ;

project : STRING
    | NUMBER
    ;

page : STRING { $$ = $1 }
    | NUMBER { $$ = $1 }
    ;

hits : NUMBER { $$ = $1 }
    ;

size : NUMBER { $$ = $1 }
    ;

%%      /*  start  of  programs  */

type Lex struct {
	fileContents []byte
	runePosition int
	currentLineNumber int
	r *strings.Reader
}

func (l *Lex) Lex(lval *yySymType) int {
// skip leading whitespace
    var c rune
	var size int
    var err error

    c, size, err = l.r.ReadRune()
    if err != nil {
    }

    // skip leading whitespace
    for unicode.IsSpace(c) && c != '\n' && err == nil {
        c, size, err = l.r.ReadRune()
    }
    if err != nil {
        return 0
    }

    if c == '\n' {
        l.currentLineNumber++
        return NL
    }

    // might be a STRING or a NUMBER
    // gobble everything up to the next WS, NL, or EOF

    var fOnlyDigits bool = true
    str := strings.Builder{}

    for ! unicode.IsSpace(c) && err == nil {
        str.WriteRune(c)

        if ! unicode.IsDigit(c) {
            fOnlyDigits = false
        }

        c, size, err = l.r.ReadRune()
    }
    if err == nil {
        l.r.UnreadRune()
    }

    val := str.String()
    if len(val) > 0 {
        lval.str = val
        if fOnlyDigits {
            lval.val, err = strconv.Atoi(val)
            return NUMBER
        } else {
            return STRING
        }
    }

    _ = size

    return 0
}

func (l *Lex) Error(s string) {
	fmt.Printf("syntax error: %s\n", s)
}

func ParseFile( filename string ) ([]Record, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return records, err
	}

    parse := yyNewParser()
    parse.Parse(&Lex{fileContents: data, r: strings.NewReader(string(data))})

    return records, nil
}


