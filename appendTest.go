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

type astNode interface {
	parent() *astNode
	context() *astClosure
	// ...
}
type astBase struct {
	up  *astNode
	ctx *astClosure
}

type astTuple struct {
	astBase
	members []astNode
}
type astClosure struct {
	astBase
	expr *astNode
}
type astClParam struct {
	astBase
}
type astClRslt struct {
	astBase
}
type astCall struct {
	astBase
	procParam *astTuple // 2-tuple of proc and param
}
type astId struct {
	astBase
	id string
}
type astNewId struct {
	astBase
	id string
}

// these should be const but not sure
var e1 = astCall{procParam: &e2} // file level ; op
var e2 = astTuple{up: &e1, members: []astNode{&e3, &e4}}
var e3 = astId{up: &e2, id: ";"} // ; operator
var e4 = astTuple{up: &e2, members: []astNode{&e5, &e6, &e7, &e8}}
var e5 = astCall{procParam: &e9}                            // define append with = call
var e9 = astTuple{up: &e5, members: []astNode{&e10, &e11}}  // e10 is =, e11 procParam
var e10 = astId{up: &e9, id: "="}                           //
var e11 = astTuple{up: &e9, members: []astNode{&e12, &e13}} // e12 is append, e3 {}
var e12 = astNewId{up: &e11, id: "append"}
var e13 = astClosure{u: &e11, expr: &e14} // the body of append must give this as ctx
var e14 = astCall{up: &e13, ctx: &e13, procParam: &e15}
var e15 = astTuple{up: &e14, ctx: &e13, members: []astNode{&e16, &e17}} // ;
var e16 = astId{up: &e15, ctx: &e13, id: "="}                           // implied `$ = body
var e17 = astTuple{up: &e15, ctx: &e13, members: []astNode{&e18, &e19}} // append code
var e18 = astClRslt{up: &e17, ctx: &e13}
var e19 = astCall{up: &e17, ctx: &e13, procParam: &e20} // the real append body
var e20 = astTuple{up: &e19, ctx: &e13, members: []astNode{&e21, &e22}}
var e21 = astId{up: &e19, ctx: &e13, id: ";"}
var e22 = astTuple{up: &e19, ctx: &e13, members: []astNode{&e23, &e24}} // append's 2 stmts
var e23 = astCall{up: &e22, ctx: &e13, procParam: &e25}
var e25 = astTuple{up: &e23, ctx: &e13, members: []astNode{&e26, &e27}}
var e26 = astId{up: &e25, ctx: &e13, id: "="}
var e27 = astTuple{up: &e25, ctx: &e13, members: []astNode{&e28, &e29}}
var e28 = astClParam{up: &e27, ctx: &e13} // append 2nd statement/expr
var e29 = astTuple{up: &e27, ctx: &e13, members: []astNode{&e30, &e31}}
var e30 = astNewId{up: &e29, ctx: &e13, id: "x"}
var e31 = astNewId{up: &e29, ctx: &e13, id: "y"}
var e24 = astCall{up: &e22, ctx: &e13, procParam: &e32} // caseP [...] ()
var e32 = astTuple{up: &e24, ctx: &e13, members: {&e33, &e34}}
var e33 = astCall{up: &e32, ctx: &e13, procParam: &e35}
var e34 = astTuple{up: &e32, ctx: &e13, members: []astNode{}} // () is a 0-tuple
var e35 = astCall{up: &e32, ctx: &e13, procParam: &e36}       // caseP [...]
var e36 = astTuple{up: &e35, ctx: &e13, members: []astNode{&e37, &e38}}
var e37 = astId{up: &e36, ctx: &e13, id: "caseP"}
var e38 = astCall{up: &e35, ctx: &e13, procParam: &e39} // [] operator = tupleToList
var e39 = astTuple{up: &e38, ctx: &e13, members: []astNode{&e40, &e41}}
var e40 = astId{up: &e39, ctx: &e13, "[]"}
var e41 = astTuple{up: &e39, ctx: &e13, members: []astNode{&e42, &e43}} // the 2 cases
var e42 = astClosure{up: &e41, ctx: &e13, expr: &e103}                  // 1st case
var e103 = astCall{up: &e42, ctx: &e42, procParam: &e104}
var e104 = astTuple{up: &e103, ctx: &e42, members: []astNode{&e105, &e106}}
var e105 = astId{up: &e104, ctx: &e42, id: "="}
var e106 = astTuple{up: &e104, ctx: &e42, members: []astNode{&e107, &e44}}
var e107 = astClRslt{up: &e106, ctx: &e42}
var e44 = astCall{up: &e42, ctx: &e42, procParam: &e45}
var e45 = astTuple{up: &e44, ctx: &e42, members: []astNode{&e46, &e47}}
var e46 = astId{up: &e45, ctx: &e42, id: ";"}
var e47 = astTuple{up: &e45, ctx: &e42, members: []astNode{&e48, &e49}} // 2 stmts 1st case
var e48 = astCall{up: &e47, ctx: &e42, procParam: &e50}                 // x=[]
var e50 = astTuple{up: &e48, ctx: &e42, members: []astNode{&e51, &e52}}
var e51 = astId{up: &e50, ctx: &e42, id: "="}
var e52 = astCall{up: &e50, ctx: &e42, procParam: &e53} // all this for []
var e53 = astTuple{up: &e52, ctx: &e42, members: []astNode{&e54, &e55}}
var e54 = astId{up: &e53, ctx: &e42, id: "x"}
var e55 = astCall{up: &e53, ctx: &e42, procParam: &e56}
var e56 = astTuple{up: &e55, ctx: &e42, members: []astNode{&e57, &e58}}
var e57 = astId{up: &e56, ctx: &e42, "[]"}
var e58 = astTuple{up: &e56, ctx: &e42, members: []astNode{}} // 0-tuple -> []
var e49 = astId{up: &e47, ctx: &e42, id: "y"}
var e43 = astClosure{up: &e41, ctx: &e13, expr: &e108} // 2nd case
var e108 = astCall{up: &e43, ctx: &e43, procParam: &e109}
var e109 = astTuple{up: &e108, ctx: &e43, members: []astNode{&e110, &e111}}
var e110 = astId{up: &e109, ctx: &e43, id: "="}
var e111 = astTuple{up: &e109, ctx: &e43, members: []astNode{&e112, &e59}}
var e112 = astClRslt{up: &e111, ctx: &e43}
var e59 = astCall{up: &e43, ctx: &e43, procParam: &e60}
var e60 = astTuple{up: &e59, ctx: &e43, members: []astNode{&e61, &e62}}
var e61 = astId{up: &e60, ctx: &e43, id: ";"}
var e62 = astTuple{up: &e60, ctx: &e43, members: []astNode{&e63, &e74}} // 2 stmts 2nd case
var e63 = astCall{up: &e62, ctx: &e43, procParam: &e64}                 // x=`hdx..
var e64 = astTuple{up: &e62, ctx: &e43, members: []astNode{&e65, &e66}}
var e65 = astId{up: &e64, ctx: &e43, id: "="}
var e66 = astTuple{up: &e64, ctx: &e43, members: []astNode{&e67, &e68}}
var e67 = astId{up: &e66, ctx: &e43, id: "x"}
var e68 = astCall{up: &e66, ctx: &e43, procParam: &e69} // `hdx +> `tlx
var e69 = astTuple{up: &e68, ctx: &e43, members: []astNode{&e70, &e71}}
var e70 = astId{up: &e69, ctx: &e43, id: "+>"}
var e71 = astTuple{up: &e69, ctx: &e43, members: []astNode{&e72, &e73}}
var e72 = astNewId{up: &e71, ctx: &e43, id: "hdx"}
var e73 = astNewId{up: &e71, ctx: &e43, id: "tlx"}
var e74 = astCall{up: &e62, ctx: &e43, procParam: &e75} // 2nd stmt
var e75 = astTuple{up: &e74, ctx: &e43, members: []astNode{&e76, &e77}}
var e76 = astId{up: &e75, ctx: &e43, id: "+>"}
var e77 = astTuple{up: &e75, ctx: &e43, members: []astNode{&e78, &e79}}
var e78 = astId{up: &e77, ctx: &e43, id: "hdx"}
var e79 = astCall{up: &e77, ctx: &e43, procParam: &e80}
var e80 = astTuple{up: &e79, ctx: &e43, members: []astNode{&e81, &e82}}
var e81 = astId{up: &e80, ctx: &e43, id: "append"}
var e82 = astTuple{up: &e80, ctx: &e43, members: []astNode{&e83, &e84}}
var e83 = astId{up: &e82, ctx: &e43, id: "tlx"}
var e84 = astId{up: &e82, ctx: &e43, id: "y"}

var e6 = astCall{up: &e4, procParam: &e85} //print( append([1 2],[3 4]))
var e85 = astTuple{up: &e6, members: []astNode{&e86, &e87}}
var e86 = astId{up: &e85, id: "print"}
var e87 = astCall{up: &e85, procParam: &e88}
var e88 = astTuple{up: &e87, members: []astNode{&e89, &e90}}
var e89 = astId{up: &e88, id: "append"}
var e90 = astTuple{up: &e88, members: []astNode{&e91, &e92}}
var e91 = astCall{up: &e90, procParam: &e93}
var e92 = astCall{up: &e90, procParam: &e94}
var e93 = astTuple{up: &e91, members: []astNode{&e95, &e96}}
var e95 = astId{up: &e93, id: "[]"}
var e96 = astTuple{up: &e93, members: []astNode{&e97, &e98}}
var e97 = astId{up: &e96, id: "1"}
var e98 = astId{up: &e96, id: "2"}
var e94 = astTuple{up: &e92, members: []astNode{&e99, &e100}}
var e99 = astId{up: &e94, id: "[]"}
var e100 = astTuple{up: &e94, members: []astNode{&e101, &e102}}
var e101 = astId{up: &e100, id: "3"}
var e102 = astId{up: &e100, id: "4"}

var e7 = astCall{up: &e4, procParam: &e113} //[1 2 3 4] = append([1 2],print(`a))
var e113 = astTuple{up: &e7, members: []astNode{&e114, &e115}}
var e114 = astId{up: &e113, id: "="}
var e115 = astTuple{up: &e113, members: []astNode{&e116, &e117}}
var e116 = astCall{up: &e115, procParam: &e118}
var e118 = astTuple{up: &e116, members: []astNode{&e119, &e120}}
var e119 = astId{up: &e118, id: "[]"}
var e120 = astTuple{up: &e118, members: []astNode{&e121, &e122, &e123, &e124}}
var e121 = astId{up: &e120, id: "1"}
var e122 = astId{up: &e120, id: "2"}
var e123 = astId{up: &e120, id: "3"}
var e124 = astId{up: &e120, id: "4"}

var e117 = astCall{up: &e115, procParam: &e125}
var e125 = astTuple{up: &e117, members: []astNode{&e126, &e127}}
var e126 = astId{up: &e125, id: "append"}
var e127 = astTuple{up: &e125, members: []astNode{&e128, &e129}} // [1 2] and print(`a)
var e128 = astCall{up: &e127, procParam: &e130}
var e130 = astTuple{up: &e128, members: []astNode{&e131, &e132}}
var e131 = astId{up: &e130, id: "[]"}
var e132 = astTuple{up: &e130, members: []astNode{&e138, &e139}}
var e138 = astId{up: &e132, id: "1"}
var e139 = astId{up: &e132, id: "2"}
var e129 = astCall{up: &e127, procParam: &e140}
var e140 = astTuple{up: &e129, members: []astNode{&e141, &e142}}
var e141 = astId{up: &e140, id: "print"}
var e142 = astNewId{up: &e140, id: "a"}

var e8 = astCall //[1 2 3 4] = append(print(`b),[3 4])
var e143 = astTuple{up: &e8, members: []astNode{&e144, &e145}}
var e144 = astId{up: &e143, id: "="}
var e145 = astTuple{up: &e143, members: []astNode{&e146, &e147}}
var e146 = astCall{up: &e145, procParam: &e148}
var e148 = astTuple{up: &e146, members: []astNode{&e149, &e150}}
var e149 = astId{up: &e148, id: "[]"}
var e150 = astTuple{up: &e148, members: []astNode{&e151, &e152, &e153, &e154}}
var e151 = astId{up: &e150, id: "1"}
var e152 = astId{up: &e150, id: "2"}
var e153 = astId{up: &e150, id: "3"}
var e154 = astId{up: &e150, id: "4"}

var e147 = astCall{up: &e145, procParam: &e155}
var e155 = astTuple{up: &e147, members: []astNode{&e156, &e157}}
var e156 = astId{up: &e155, id: "append"}
var e157 = astTuple{up: &e155, members: []astNode{&e159, &e158}} // print(`b) and [3 4]
var e159 = astCall{up: &e157, procParam: &e170}
var e170 = astTuple{up: &e159, members: []astNode{&e171, &e172}}
var e171 = astId{up: &e170, id: "print"}
var e172 = astNewId{up: &e170, id: "b"}
var e158 = astCall{up: &e157, procParam: &e160}
var e160 = astTuple{up: &e158, members: []astNode{&e161, &e162}}
var e161 = astId{up: &e160, id: "[]"}
var e162 = astTuple{up: &e160, members: []astNode{&e168, &e169}}
var e168 = astId{up: &e162, id: "3"}
var e169 = astId{up: &e162, id: "4"}
