package main

/*
The program we hand convert to AST and interpret is:

    `append = {
        $ = (`x,`y); # 2 input lists
        caseP [
            { x:[]; y }
            { x = `hdx +> `tlx;
              hdx +> append(tlx,y) }
        ] ()
    };

    print( append([1 2],[3 4])); # [1 2 3 4]
    [1 2 3 4] = append([1 2],print(`a)); # [3 4] -- print returns its argument
    [1 2 3 4] = append(print(`b),[3 4]); # [1 2]

We have a marsupial AST. From the doc, the nodes relevant to us:
 - Tuple. The 0-tuple is wombat's unit:Unit.
 - Closure. Includes inner.
 - ClosureParam. From $ operator in wombat.
 - ClosureResult. From `$ operator in wombat.
 - Call. The 2-tuple parameter has procedure and parameter.
 - Identifier. A name that is not an operator or currently active suboperator.
 - NewIdentifier. Precede with backquote. Cannot be redeclared in same scope.

Names declared before this fragment are:
 - 1, 2, 3 and 4
 - print
 - caseP
plus operator operations:
 - ';' takes a tuple, unifying last element with result
 - '=' takes a 2-tuple and unifies them (returning that result, but not used here)
 - '[]'=tupleToList - a convertsTo
 - +> prepend

*/
//import "fmt"

type astNode interface {
	parent() astNode
	context() *astClosure
	pp()
	// ...
}

type astTuple struct {
	up      astNode
	ctx     *astClosure
	members []astNode
}

func (x astTuple) parent() astNode      { return x.up }
func (x astTuple) context() *astClosure { return x.ctx }
func (x astTuple) pp() {
	print(" (")
	for i := 0; i < len(x.members); i += 1 {
		//if x.members[i].parent().(astTuple) != x {
		//	panic("tuple member wrong parent")
		//}
		x.members[i].pp()
		if i != len(x.members)-1 {
			print(",")
		}
	}
	print(") ")
}

type astClosure struct {
	up   astNode
	ctx  *astClosure
	expr astNode
}

func (x astClosure) parent() astNode      { return x.up }
func (x astClosure) context() *astClosure { return x.ctx }
func (x astClosure) pp()                  { print(" {"); (x.expr).pp(); print("} ") }

type astClParam struct {
	up  astNode
	ctx *astClosure
}

func (x astClParam) parent() astNode      { return x.up }
func (x astClParam) context() *astClosure { return x.ctx }
func (x astClParam) pp()                  { print(" $ ") }

type astClRslt struct {
	up  astNode
	ctx *astClosure
}

func (x astClRslt) parent() astNode      { return x.up }
func (x astClRslt) context() *astClosure { return x.ctx }
func (x astClRslt) pp()                  { print(" `$ ") }

type astId struct {
	up  astNode
	ctx *astClosure
	id  string
}

func (x astId) parent() astNode      { return x.up }
func (x astId) context() *astClosure { return x.ctx }
func (x astId) pp()                  { print(" ", x.id, " ") }

type astNewId struct {
	up  astNode
	ctx *astClosure
	id  string
}

func (x astNewId) parent() astNode      { return x.up }
func (x astNewId) context() *astClosure { return x.ctx }
func (x astNewId) pp()                  { print(" `", x.id, " ") }

type astCall struct {
	up        astNode
	ctx       *astClosure
	procParam *astTuple // 2-tuple of proc and param
}

func (x astCall) parent() astNode      { return x.up }
func (x astCall) context() *astClosure { return x.ctx }
func (x astCall) pp() {
	// procParam should be a 2-tuple
	if len(x.procParam.members) != 2 {
		panic("Call procParam should be 2-tuple")
	}
	switch f := x.procParam.members[0]; f.(type) {
	case *astId:
		id := f.(*astId).id
		switch id {
		case "=", "+>": // infix
			p := x.procParam.members[1].(*astTuple) // should be 2-tuple
			if len(p.members) != 2 {
				panic(id + " must have 2 parameters")
			}
			p.members[0].pp()
			print(id + " (")
			p.members[1].pp()
			print(") ")
		case ";":
			q := x.procParam.members[1].(*astTuple)
			for i := 0; i < len(q.members); i++ {
				print("\n  ")
				q.members[i].pp()
				if i != len(q.members)-1 {
					print(";")
				}
			}
		case "[]":
			r := x.procParam.members[1].(*astTuple)
			print("[")
			for j := 0; j < len(r.members); j++ {
				print("(")
				r.members[j].pp()
				print(")")
				if j != len(r.members)-1 {
					print(" ")
				}
			}
			print("]")
		default: // ordinary function call
			print(id)
			print("(")
			x.procParam.members[1].pp()
			print(")")
		}
	default:
		print("(")
		//print(t)
		f.pp()
		print(")(")
		x.procParam.members[1].pp()
		print(")")
	}
}

// these should be const but not sure
var e1 astCall  //= astCall{procParam: &e2} // file level ; op
var e2 astTuple //= astTuple{up: &e1, members: []astNode{&e3, &e4}}
var e3 astId    //= astId{up: &e2, id: ";"} // ; operator
var e4 astTuple //= astTuple{up: &e2, members: []astNode{&e5, &e6, &e7, &e8}}
var e5 astCall  //=var e1 astCall = astCall{procParam: &e2} // file level ; op

var e9 astTuple    //= astTuple{up: &e5, members: []astNode{&e10, &e11}}  // e10 is =, e11 procParam
var e10 astId      //= astId{up: &e9, id: "="}                           //
var e11 astTuple   //= astTuple{up: &e9, members: []astNode{&e12, &e13}} // e12 is append, e3 {}
var e12 astNewId   //= astNewId{up: &e11, id: "append"}
var e13 astClosure //= astClosure{up: &e11, expr: &e14} // the body of append must give this as ctx
var e14 astCall    //= astCall{up: &e13, ctx: &e13, procParam: &e15}
var e15 astTuple   //= astTuple{up: &e14, ctx: &e13, members: []astNode{&e16, &e17}} // ;
var e16 astId      //= astId{up: &e15, ctx: &e13, id: "="}                           // implied `$ = body
var e17 astTuple   //= astTuple{up: &e15, ctx: &e13, members: []astNode{&e18, &e19}} // append code
var e18 astClRslt  //= astClRslt{up: &e17, ctx: &e13}
var e19 astCall    //= astCall{up: &e17, ctx: &e13, procParam: &e20} // the real append body
var e20 astTuple   //= astTuple{up: &e19, ctx: &e13, members: []astNode{&e21, &e22}}
var e21 astId      //= astId{up: &e19, ctx: &e13, id: ";"}
var e22 astTuple   //= astTuple{up: &e19, ctx: &e13, members: []astNode{&e23, &e24}} // append's 2 stmts
var e23 astCall    //= astCall{up: &e22, ctx: &e13, procParam: &e25}
var e25 astTuple   //= astTuple{up: &e23, ctx: &e13, members: []astNode{&e26, &e27}}
var e26 astId      //= astId{up: &e25, ctx: &e13, id: "="}
var e27 astTuple   //= astTuple{up: &e25, ctx: &e13, members: []astNode{&e28, &e29}}
var e28 astClParam //= astClParam{up: &e27, ctx: &e13} // append 2nd statement/expr
var e29 astTuple   //= astTuple{up: &e27, ctx: &e13, members: []astNode{&e30, &e31}}
var e30 astNewId   //= astNewId{up: &e29, ctx: &e13, id: "x"}
var e31 astNewId   //= astNewId{up: &e29, ctx: &e13, id: "y"}
var e24 astCall    //= astCall{up: &e22, ctx: &e13, procParam: &e32} // caseP [...] ()
var e32 astTuple   //= astTuple{up: &e24, ctx: &e13, members: []astNode{&e33, &e34}}
var e33 astCall    //= astCall{up: &e32, ctx: &e13, procParam: &e36}
var e34 astTuple   //= astTuple{up: &e32, ctx: &e13, members: []astNode{}} // () is a 0-tuple
//var e173 astTuple = astTuple{up: &e33, ctx: &e13, members:[]astNode{&e37
//var e35 astCall = astCall{up: &e32, ctx: &e13, procParam: &e36}       // caseP [...]
var e36 astTuple   //= astTuple{up: &e33, ctx: &e13, members: []astNode{&e37, &e38}}
var e37 astId      //= astId{up: &e36, ctx: &e13, id: "caseP"}
var e38 astCall    //= astCall{up: &e36, ctx: &e13, procParam: &e39} // [] operator = tupleToList
var e39 astTuple   //= astTuple{up: &e38, ctx: &e13, members: []astNode{&e40, &e41}}
var e40 astId      //= astId{up: &e39, ctx: &e13, id: "[]"}
var e41 astTuple   //= astTuple{up: &e39, ctx: &e13, members: []astNode{&e42, &e43}} // the 2 cases
var e42 astClosure //= astClosure{up: &e41, ctx: &e13, expr: &e103}                  // 1st case
var e103 astCall   //= astCall{up: &e42, ctx: &e42, procParam: &e104}
var e104 astTuple  //= astTuple{up: &e103, ctx: &e42, members: []astNode{&e105, &e106}}
var e105 astId     //= astId{up: &e104, ctx: &e42, id: "="}
var e106 astTuple  //= astTuple{up: &e104, ctx: &e42, members: []astNode{&e107, &e44}}
var e107 astClRslt //= astClRslt{up: &e106, ctx: &e42}
var e44 astCall    //= astCall{up: &e42, ctx: &e42, procParam: &e45}
var e45 astTuple   //= astTuple{up: &e44, ctx: &e42, members: []astNode{&e46, &e47}}
var e46 astId      //= astId{up: &e45, ctx: &e42, id: ";"}
var e47 astTuple   //= astTuple{up: &e45, ctx: &e42, members: []astNode{&e48, &e49}} // 2 stmts 1st case
var e48 astCall    //= astCall{up: &e47, ctx: &e42, procParam: &e50}                 // x=[]
var e50 astTuple   //= astTuple{up: &e48, ctx: &e42, members: []astNode{&e51, &e174}}
var e51 astId      //= astId{up: &e50, ctx: &e42, id: "="}
var e174 astTuple  //= astTuple{up: &e50, ctx: &e42, members: []astNode{&e54, &e55}}
var e52 astCall    //= astCall{up: &e50, ctx: &e42, procParam: &e53} // all this for []
var e53 astTuple   //= astTuple{up: &e52, ctx: &e42, members: []astNode{&e54, &e55}}
var e54 astId      //= astId{up: &e174, ctx: &e42, id: "x"}
var e55 astCall    //= astCall{up: &e174, ctx: &e42, procParam: &e56}
var e56 astTuple   //= astTuple{up: &e55, ctx: &e42, members: []astNode{&e57, &e58}}
var e57 astId      //= astId{up: &e56, ctx: &e42, id: "[]"}
//r: unexpected x, expecting semicolon, newline, or }
var e58 astTuple   //= astTuple{up: &e56, ctx: &e42, members: []astNode{}} // 0-tuple -> []
var e49 astId      //= astId{up: &e47, ctx: &e42, id: "y"}
var e43 astClosure //= astClosure{up: &e41, ctx: &e13, expr: &e108} // 2nd case
var e108 astCall   //= astCall{up: &e43, ctx: &e43, procParam: &e109}
var e109 astTuple  //= astTuple{up: &e108, ctx: &e43, members: []astNode{&e110, &e111}}
var e110 astId     //= astId{up: &e109, ctx: &e43, id: "="}
var e111 astTuple  //= astTuple{up: &e109, ctx: &e43, members: []astNode{&e112, &e59}}
var e112 astClRslt //= astClRslt{up: &e111, ctx: &e43}
var e59 astCall    //= astCall{up: &e43, ctx: &e43, procParam: &e60}
var e60 astTuple   //= astTuple{up: &e59, ctx: &e43, members: []astNode{&e61, &e62}}
var e61 astId      //= astId{up: &e60, ctx: &e43, id: ";"}
var e62 astTuple   //= astTuple{up: &e60, ctx: &e43, members: []astNode{&e63, &e74}} // 2 stmts 2nd case
var e63 astCall    //= astCall{up: &e62, ctx: &e43, procParam: &e64}                 // x=`hdx..
var e64 astTuple   //= astTuple{up: &e62, ctx: &e43, members: []astNode{&e65, &e66}}
var e65 astId      //= astId{up: &e64, ctx: &e43, id: "="}
var e66 astTuple   //= astTuple{up: &e64, ctx: &e43, members: []astNode{&e67, &e68}}
var e67 astId      //= astId{up: &e66, ctx: &e43, id: "x"}
//r: unexpected x, expecting semicolon, newline, or }r
var e68 astCall  //= astCall{up: &e66, ctx: &e43, procParam: &e69} // `hdx +> `tlx
var e69 astTuple //= astTuple{up: &e68, ctx: &e43, members: []astNode{&e70, &e71}}
var e70 astId    //= astId{up: &e69, ctx: &e43, id: "+>"}
var e71 astTuple //= astTuple{up: &e69, ctx: &e43, members: []astNode{&e72, &e73}}
var e72 astNewId //= astNewId{up: &e71, ctx: &e43, id: "hdx"}
var e73 astNewId //= astNewId{up: &e71, ctx: &e43, id: "tlx"}
var e74 astCall  //= astCall{up: &e62, ctx: &e43, procParam: &e75} // 2nd stmt
var e75 astTuple //= astTuple{up: &e74, ctx: &e43, members: []astNode{&e76, &e77}}
var e76 astId    //= astId{up: &e75, ctx: &e43, id: "+>"}
var e77 astTuple //= astTuple{up: &e75, ctx: &e43, members: []astNode{&e78, &e79}}
var e78 astId    //= astId{up: &e77, ctx: &e43, id: "hdx"}
var e79 astCall  //= astCall{up: &e77, ctx: &e43, procParam: &e80}
var e80 astTuple //= astTuple{up: &e79, ctx: &e43, members: []astNode{&e81, &e82}}
var e81 astId    //= astId{up: &e80, ctx: &e43, id: "append"}
var e82 astTuple //= astTuple{up: &e80, ctx: &e43, members: []astNode{&e83, &e84}}
var e83 astId    //= astId{up: &e82, ctx: &e43, id: "tlx"}
var e84 astId    //= astId{up: &e82, ctx: &e43, id: "y"}

var e6 astCall    //= astCall{up: &e4, procParam: &e85} //print( append([1 2],[3 4]))
var e85 astTuple  //= astTuple{up: &e6, members: []astNode{&e86, &e87}}
var e86 astId     //= astId{up: &e85, id: "print"}
var e87 astCall   //= astCall{up: &e85, procParam: &e88}
var e88 astTuple  //= astTuple{up: &e87, members: []astNode{&e89, &e90}}
var e89 astId     //= astId{up: &e88, id: "append"}
var e90 astTuple  //= astTuple{up: &e88, members: []astNode{&e91, &e92}}
var e91 astCall   //= astCall{up: &e90, procParam: &e93}
var e92 astCall   //= astCall{up: &e90, procParam: &e94}
var e93 astTuple  //= astTuple{up: &e91, members: []astNode{&e95, &e96}}
var e95 astId     //= astId{up: &e93, id: "[]"}
var e96 astTuple  //= astTuple{up: &e93, members: []astNode{&e97, &e98}}
var e97 astId     //= astId{up: &e96, id: "1"}
var e98 astId     //= astId{up: &e96, id: "2"}
var e94 astTuple  //= astTuple{up: &e92, members: []astNode{&e99, &e100}}
var e99 astId     //= astId{up: &e94, id: "[]"}
var e100 astTuple //= astTuple{up: &e94, members: []astNode{&e101, &e102}}
var e101 astId    //= astId{up: &e100, id: "3"}
var e102 astId    //= astId{up: &e100, id: "4"}

var e7 astCall    //= astCall{up: &e4, procParam: &e113} //[1 2 3 4] = append([1 2],print(`a))
var e113 astTuple //= astTuple{up: &e7, members: []astNode{&e114, &e115}}
var e114 astId    //= astId{up: &e113, id: "="}
var e115 astTuple //= astTuple{up: &e113, members: []astNode{&e116, &e117}}
var e116 astCall  //= astCall{up: &e115, procParam: &e118}
var e118 astTuple //= astTuple{up: &e116, members: []astNode{&e119, &e120}}
var e119 astId    //= astId{up: &e118, id: "[]"}
var e120 astTuple //= astTuple{up: &e118, members: []astNode{&e121, &e122, &e123, &e124}}
var e121 astId    //= astId{up: &e120, id: "1"}
var e122 astId    //= astId{up: &e120, id: "2"}
var e123 astId    //= astId{up: &e120, id: "3"}
var e124 astId    //= astId{up: &e120, id: "4"}

var e117 astCall  //= astCall{up: &e115, procParam: &e125}
var e125 astTuple //= astTuple{up: &e117, members: []astNode{&e126, &e127}}
var e126 astId    //= astId{up: &e125, id: "append"}
var e127 astTuple //= astTuple{up: &e125, members: []astNode{&e128, &e129}} // [1 2] and print(`a)
var e128 astCall  //= astCall{up: &e127, procParam: &e130}
var e130 astTuple //= astTuple{up: &e128, members: []astNode{&e131, &e132}}
var e131 astId    //= astId{up: &e130, id: "[]"}
var e132 astTuple //= astTuple{up: &e130, members: []astNode{&e138, &e139}}
var e138 astId    //= astId{up: &e132, id: "1"}
var e139 astId    //= astId{up: &e132, id: "2"}
//r: unexpected x, expecting semicolon, newline, or }
var e129 astCall  //= astCall{up: &e127, procParam: &e140}
var e140 astTuple //= astTuple{up: &e129, members: []astNode{&e141, &e142}}
var e141 astId    //= astId{up: &e140, id: "print"}
var e142 astNewId //= astNewId{up: &e140, id: "a"}

var e8 astCall    //= astCall{up: &e4, procParam: &e143} //[1 2 3 4] = append(print(`b),[3 4])
var e143 astTuple //= astTuple{up: &e8, members: []astNode{&e144, &e145}}
var e144 astId    //= astId{up: &e143, id: "="}
var e145 astTuple //= astTuple{up: &e143, members: []astNode{&e146, &e147}}
var e146 astCall  //= astCall{up: &e145, procParam: &e148}
var e148 astTuple //= astTuple{up: &e146, members: []astNode{&e149, &e150}}
var e149 astId    //= astId{up: &e148, id: "[]"}
var e150 astTuple //= astTuple{up: &e148, members: []astNode{&e151, &e152, &e153, &e154}}
var e151 astId    //= astId{up: &e150, id: "1"}
var e152 astId    //= astId{up: &e150, id: "2"}
var e153 astId    //= astId{up: &e150, id: "3"}
var e154 astId    //= astId{up: &e150, id: "4"}

var e147 astCall  //= astCall{up: &e145, procParam: &e155}
var e155 astTuple //= astTuple{up: &e147, members: []astNode{&e156, &e157}}
var e156 astId    //= astId{up: &e155, id: "append"}
var e157 astTuple //= astTuple{up: &e155, members: []astNode{&e159, &e158}} // print(`b) and [3 4]
var e159 astCall  //= astCall{up: &e157, procParam: &e170}
var e170 astTuple //= astTuple{up: &e159, members: []astNode{&e171, &e172}}
var e171 astId    //= astId{up: &e170, id: "print"}
var e172 astNewId //= astNewId{up: &e170, id: "b"}
var e158 astCall  //= astCall{up: &e157, procParam: &e160}
var e160 astTuple //= astTuple{up: &e158, members: []astNode{&e161, &e162}}
var e161 astId    //= astId{up: &e160, id: "[]"}
var e162 astTuple //= astTuple{up: &e160, members: []astNode{&e168, &e169}}
var e168 astId    //= astId{up: &e162, id: "3"}
var e169 astId    //= astId{up: &e162, id: "4"}

func init() {
	e1 = astCall{procParam: &e2} // file level ; op
	e2 = astTuple{up: &e1, members: []astNode{&e3, &e4}}
	e3 = astId{up: &e2, id: ";"} // ; operator
	e4 = astTuple{up: &e2, members: []astNode{&e5, &e6, &e7, &e8}}
	e5 = astCall{procParam: &e9}                            // define append with = call
	e9 = astTuple{up: &e5, members: []astNode{&e10, &e11}}  // e10 is =, e11 procParam
	e10 = astId{up: &e9, id: "="}                           //
	e11 = astTuple{up: &e9, members: []astNode{&e12, &e13}} // e12 is append, e3 {}
	e12 = astNewId{up: &e11, id: "append"}
	e13 = astClosure{up: &e11, expr: &e14} // the body of append must give this as ctx
	e14 = astCall{up: &e13, ctx: &e13, procParam: &e15}
	e15 = astTuple{up: &e14, ctx: &e13, members: []astNode{&e16, &e17}} // ;
	e16 = astId{up: &e15, ctx: &e13, id: "="}                           // implied `$ = body
	e17 = astTuple{up: &e15, ctx: &e13, members: []astNode{&e18, &e19}} // append code
	e18 = astClRslt{up: &e17, ctx: &e13}
	e19 = astCall{up: &e17, ctx: &e13, procParam: &e20} // the real append body
	e20 = astTuple{up: &e19, ctx: &e13, members: []astNode{&e21, &e22}}
	e21 = astId{up: &e19, ctx: &e13, id: ";"}
	e22 = astTuple{up: &e19, ctx: &e13, members: []astNode{&e23, &e24}} // append's 2 stmts
	e23 = astCall{up: &e22, ctx: &e13, procParam: &e25}
	e25 = astTuple{up: &e23, ctx: &e13, members: []astNode{&e26, &e27}}
	e26 = astId{up: &e25, ctx: &e13, id: "="}
	e27 = astTuple{up: &e25, ctx: &e13, members: []astNode{&e28, &e29}}
	e28 = astClParam{up: &e27, ctx: &e13} // append 2nd statement/expr
	e29 = astTuple{up: &e27, ctx: &e13, members: []astNode{&e30, &e31}}
	e30 = astNewId{up: &e29, ctx: &e13, id: "x"}
	e31 = astNewId{up: &e29, ctx: &e13, id: "y"}
	e24 = astCall{up: &e22, ctx: &e13, procParam: &e32} // caseP [...] ()
	e32 = astTuple{up: &e24, ctx: &e13, members: []astNode{&e33, &e34}}
	e33 = astCall{up: &e32, ctx: &e13, procParam: &e36}       // caseP [...]
	e34 = astTuple{up: &e32, ctx: &e13, members: []astNode{}} // () is a 0-tuple
	//e173 = astTuple{up: &e33, ctx: &e13, members:[]astNode{&e37
	//e35 = astCall{up: &e32, ctx: &e13, procParam: &e36}
	e36 = astTuple{up: &e33, ctx: &e13, members: []astNode{&e37, &e38}}
	e37 = astId{up: &e36, ctx: &e13, id: "caseP"}
	e38 = astCall{up: &e36, ctx: &e13, procParam: &e39} // [] operator = tupleToList
	e39 = astTuple{up: &e38, ctx: &e13, members: []astNode{&e40, &e41}}
	e40 = astId{up: &e39, ctx: &e13, id: "[]"}
	e41 = astTuple{up: &e39, ctx: &e13, members: []astNode{&e42, &e43}} // the 2 cases
	e42 = astClosure{up: &e41, ctx: &e13, expr: &e103}                  // 1st case
	e103 = astCall{up: &e42, ctx: &e42, procParam: &e104}
	e104 = astTuple{up: &e103, ctx: &e42, members: []astNode{&e105, &e106}}
	e105 = astId{up: &e104, ctx: &e42, id: "="}
	e106 = astTuple{up: &e104, ctx: &e42, members: []astNode{&e107, &e44}}
	e107 = astClRslt{up: &e106, ctx: &e42}
	e44 = astCall{up: &e42, ctx: &e42, procParam: &e45}
	e45 = astTuple{up: &e44, ctx: &e42, members: []astNode{&e46, &e47}}
	e46 = astId{up: &e45, ctx: &e42, id: ";"}
	e47 = astTuple{up: &e45, ctx: &e42, members: []astNode{&e48, &e49}} // 2 stmts 1st case
	e48 = astCall{up: &e47, ctx: &e42, procParam: &e50}                 // x=[]
	//	e50 = astTuple{up: &e48, ctx: &e42, members: []astNode{&e51, &e52}}
	e50 = astTuple{up: &e48, ctx: &e42, members: []astNode{&e51, &e53}}
	e51 = astId{up: &e50, ctx: &e42, id: "="}
	//	e174 = astTuple{up: &e50, ctx: &e42, members: []astNode{&e52, &e49}}
	//	e52 = astCall{up: &e50, ctx: &e42, procParam: &e53} // all this for []
	//	e52 = astCall{up: &e174, ctx: &e42, procParam: &e53}
	e53 = astTuple{up: &e52, ctx: &e42, members: []astNode{&e54, &e55}}
	e54 = astId{up: &e53, ctx: &e42, id: "x"}
	e55 = astCall{up: &e53, ctx: &e42, procParam: &e56}
	e56 = astTuple{up: &e55, ctx: &e42, members: []astNode{&e57, &e58}}
	e57 = astId{up: &e56, ctx: &e42, id: "[]"}
	//r: unexpected x, expecting semicolon, newline, or }
	e58 = astTuple{up: &e56, ctx: &e42, members: []astNode{}} // 0-tuple -> []
	e49 = astId{up: &e47, ctx: &e42, id: "y"}
	e43 = astClosure{up: &e41, ctx: &e13, expr: &e108} // 2nd case
	e108 = astCall{up: &e43, ctx: &e43, procParam: &e109}
	e109 = astTuple{up: &e108, ctx: &e43, members: []astNode{&e110, &e111}}
	e110 = astId{up: &e109, ctx: &e43, id: "="}
	e111 = astTuple{up: &e109, ctx: &e43, members: []astNode{&e112, &e59}}
	e112 = astClRslt{up: &e111, ctx: &e43}
	e59 = astCall{up: &e43, ctx: &e43, procParam: &e60}
	e60 = astTuple{up: &e59, ctx: &e43, members: []astNode{&e61, &e62}}
	e61 = astId{up: &e60, ctx: &e43, id: ";"}
	e62 = astTuple{up: &e60, ctx: &e43, members: []astNode{&e63, &e74}} // 2 stmts 2nd case
	e63 = astCall{up: &e62, ctx: &e43, procParam: &e64}                 // x=`hdx..
	e64 = astTuple{up: &e62, ctx: &e43, members: []astNode{&e65, &e66}}
	e65 = astId{up: &e64, ctx: &e43, id: "="}
	e66 = astTuple{up: &e64, ctx: &e43, members: []astNode{&e67, &e68}}
	e67 = astId{up: &e66, ctx: &e43, id: "x"}
	//r: unexpected x, expecting semicolon, newline, or }r
	e68 = astCall{up: &e66, ctx: &e43, procParam: &e69} // `hdx +> `tlx
	e69 = astTuple{up: &e68, ctx: &e43, members: []astNode{&e70, &e71}}
	e70 = astId{up: &e69, ctx: &e43, id: "+>"}
	e71 = astTuple{up: &e69, ctx: &e43, members: []astNode{&e72, &e73}}
	e72 = astNewId{up: &e71, ctx: &e43, id: "hdx"}
	e73 = astNewId{up: &e71, ctx: &e43, id: "tlx"}
	e74 = astCall{up: &e62, ctx: &e43, procParam: &e75} // 2nd stmt
	e75 = astTuple{up: &e74, ctx: &e43, members: []astNode{&e76, &e77}}
	e76 = astId{up: &e75, ctx: &e43, id: "+>"}
	e77 = astTuple{up: &e75, ctx: &e43, members: []astNode{&e78, &e79}}
	e78 = astId{up: &e77, ctx: &e43, id: "hdx"}
	e79 = astCall{up: &e77, ctx: &e43, procParam: &e80}
	e80 = astTuple{up: &e79, ctx: &e43, members: []astNode{&e81, &e82}}
	e81 = astId{up: &e80, ctx: &e43, id: "append"}
	e82 = astTuple{up: &e80, ctx: &e43, members: []astNode{&e83, &e84}}
	e83 = astId{up: &e82, ctx: &e43, id: "tlx"}
	e84 = astId{up: &e82, ctx: &e43, id: "y"}

	e6 = astCall{up: &e4, procParam: &e85} //print( append([1 2],[3 4]))
	e85 = astTuple{up: &e6, members: []astNode{&e86, &e87}}
	e86 = astId{up: &e85, id: "print"}
	e87 = astCall{up: &e85, procParam: &e88}
	e88 = astTuple{up: &e87, members: []astNode{&e89, &e90}}
	e89 = astId{up: &e88, id: "append"}
	e90 = astTuple{up: &e88, members: []astNode{&e91, &e92}}
	e91 = astCall{up: &e90, procParam: &e93}
	e92 = astCall{up: &e90, procParam: &e94}
	e93 = astTuple{up: &e91, members: []astNode{&e95, &e96}}
	e95 = astId{up: &e93, id: "[]"}
	e96 = astTuple{up: &e93, members: []astNode{&e97, &e98}}
	e97 = astId{up: &e96, id: "1"}
	e98 = astId{up: &e96, id: "2"}
	e94 = astTuple{up: &e92, members: []astNode{&e99, &e100}}
	e99 = astId{up: &e94, id: "[]"}
	e100 = astTuple{up: &e94, members: []astNode{&e101, &e102}}
	e101 = astId{up: &e100, id: "3"}
	e102 = astId{up: &e100, id: "4"}

	e7 = astCall{up: &e4, procParam: &e113} //[1 2 3 4] = append([1 2],print(`a))
	e113 = astTuple{up: &e7, members: []astNode{&e114, &e115}}
	e114 = astId{up: &e113, id: "="}
	e115 = astTuple{up: &e113, members: []astNode{&e116, &e117}}
	e116 = astCall{up: &e115, procParam: &e118}
	e118 = astTuple{up: &e116, members: []astNode{&e119, &e120}}
	e119 = astId{up: &e118, id: "[]"}
	e120 = astTuple{up: &e118, members: []astNode{&e121, &e122, &e123, &e124}}
	e121 = astId{up: &e120, id: "1"}
	e122 = astId{up: &e120, id: "2"}
	e123 = astId{up: &e120, id: "3"}
	e124 = astId{up: &e120, id: "4"}

	e117 = astCall{up: &e115, procParam: &e125}
	e125 = astTuple{up: &e117, members: []astNode{&e126, &e127}}
	e126 = astId{up: &e125, id: "append"}
	e127 = astTuple{up: &e125, members: []astNode{&e128, &e129}} // [1 2] and print(`a)
	e128 = astCall{up: &e127, procParam: &e130}
	e130 = astTuple{up: &e128, members: []astNode{&e131, &e132}}
	e131 = astId{up: &e130, id: "[]"}
	e132 = astTuple{up: &e130, members: []astNode{&e138, &e139}}
	e138 = astId{up: &e132, id: "1"}
	e139 = astId{up: &e132, id: "2"}
	//r: unexpected x, expecting semicolon, newline, or }
	e129 = astCall{up: &e127, procParam: &e140}
	e140 = astTuple{up: &e129, members: []astNode{&e141, &e142}}
	e141 = astId{up: &e140, id: "print"}
	e142 = astNewId{up: &e140, id: "a"}

	e8 = astCall{up: &e4, procParam: &e143} //[1 2 3 4] = append(print(`b),[3 4])
	e143 = astTuple{up: &e8, members: []astNode{&e144, &e145}}
	e144 = astId{up: &e143, id: "="}
	e145 = astTuple{up: &e143, members: []astNode{&e146, &e147}}
	e146 = astCall{up: &e145, procParam: &e148}
	e148 = astTuple{up: &e146, members: []astNode{&e149, &e150}}
	e149 = astId{up: &e148, id: "[]"}
	e150 = astTuple{up: &e148, members: []astNode{&e151, &e152, &e153, &e154}}
	e151 = astId{up: &e150, id: "1"}
	e152 = astId{up: &e150, id: "2"}
	e153 = astId{up: &e150, id: "3"}
	e154 = astId{up: &e150, id: "4"}

	e147 = astCall{up: &e145, procParam: &e155}
	e155 = astTuple{up: &e147, members: []astNode{&e156, &e157}}
	e156 = astId{up: &e155, id: "append"}
	e157 = astTuple{up: &e155, members: []astNode{&e159, &e158}} // print(`b) and [3 4]
	e159 = astCall{up: &e157, procParam: &e170}
	e170 = astTuple{up: &e159, members: []astNode{&e171, &e172}}
	e171 = astId{up: &e170, id: "print"}
	e172 = astNewId{up: &e170, id: "b"}
	e158 = astCall{up: &e157, procParam: &e160}
	e160 = astTuple{up: &e158, members: []astNode{&e161, &e162}}
	e161 = astId{up: &e160, id: "[]"}
	e162 = astTuple{up: &e160, members: []astNode{&e168, &e169}}
	e168 = astId{up: &e162, id: "3"}
	e169 = astId{up: &e162, id: "4"}
}

func main() { // pretty print ast
	e1.pp()
	println()
}
