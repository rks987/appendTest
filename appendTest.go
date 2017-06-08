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
 - ClosureParam. From $ operator in wo

,&e5,&e6}} // the tuple of statements
mbat.
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
var e1 = astCall{procParam:&e2}     // file level ; op
var e2 = astTuple{up:&e1,members:[]astNode{&e3,&e4}}
var e3 = astId{up:&e2,id:";"}       // ; operator
var e4 = astTuple{up:&e2,members:[]astNode{&e5,&e6,&e7,&e8}}
var e5 = astCall{procParam:&e9}     // define append with = call
var e9 = astTuple{up:&e5,members:[]astNode{&e10,&e11}}  // e10 is =, e11 procParam
var e10 = astId{up:&e9,id:"="} // 
var e11 = astTuple{up:&e9,members:[]astNode{&e12,&e13}} // e12 is append, e3 {}
var e12 = astNewId{up:&e11,id:"append"}
var e13 = astClosure{u:&e11,expr:&e14} // the body of append must give this as ctx
var e14 = astCall{up:&e13,ctx:&e13,procParam:&e15}
var e15 = astTuple{up:&e14,ctx:&e13,members:[]astNode{&e16,&e17}} // ;
var e16 = astId{up:&e15,ctx:&e13,id:"="}   // implied `$ = body
var e17 = astTuple{up:&e15,ctx:&e13,members:[]astNode{&e18,&e19}} // append code
var e18 = astClRslt{up:&e17,ctx:&e13}
var e19 = astCall{up:&e17,ctx:&e13,procParam:&e20} // the real append body
var e20 = astTuple{up:&e19,ctx:&e13,members:[]astNode{&e21,&e22}}
var e21 = astId{up:&e19,ctx:&e13,id:";"}
var e22 = astTuple{up:&e19,ctx:&e13,members:[]astNode{&e23,&e24}} // append's 2 stmts
var e23 = astCall{up:&e22,ctx:&e13,procParam:&e25}
var e25 = astTuple{up:&e23,ctx:&e13,members:[]astNode{&e26,&e27}}
var e26 = astId{up:&e25,ctx:&e13,id:"="}
var e27 = astTuple{up:&e25,ctx:&e13,members:[]astNode{&e28,&e29}}
var e28 = astClParam{up:&e27,ctx:&e13}  // append 2nd statement/expr
var e29 = astTuple{up:&e27,ctx:&e13,members:[]astNode{&e30,&e31}}
var e30 = astNewId{up:&e29,ctx:&e13,id:"x"}
var e31 = astNewId{up:&e29,ctx:&e13,id:"y"}
var e24 = astCall{up:&e22,ctx:&e13,procParam:&e32}  // caseP [...] ()
var e32 = astTuple{up:&e24,ctx:&e13,members:{&e33,&e34}}
var e33 = astCall{up:&e32,ctx:&e13,procParam:&e35}
var e34 = astTuple{up:&e32,ctx:&e13,members:[]astNode{}}  // () is a 0-tuple
var e35 = astCall{up:&e32,ctx:&e13,procParam:&e36}  // caseP [...]
var e36 = astTuple{up:&e35,ctx:&e13,members:[]astNode{&e37,&e38}}
var e37 = astId{up:&e36,ctx:&e13,id:"caseP"}
var e38 = astCall{up:&e35,ctx:&e13,procParam:&e39} // [] operator = tupleToList
var e39 = astTuple{up:&e38,ctx:&e13,members:[]astNode{&e40,&e41}}
var e40 = astId{up:&e39,ctx:&e13,"[]"}
var e41 = astTuple{up:&e39,ctx:&e13,members:[]astNode{&e42,&e43}} // the 2 cases
var e42 = astClosure{up:&e41,ctx:&e13,expr:&e44} // 1st case
var e44 = astCall{up:&e42,ctx:&e42,procParam:&e45}
var e45 = astTuple{up:&e44,ctx:&e42,members:[]astNode{&e46,&e47}}
var e46 = astId{up:&e45,ctx:&e42,id:";"}
var e47 = astTuple{up:&e45,ctx:&e42,members:[]astNode{&e48,&e49}} // 2 stmts 1st case
var e48 = astCall{up:&e47,ctx:&e42,procParam:&e50}  // x=[]
var e50 = astTuple{up:&e48,ctx:&e42,members:[]astNode{&e51,&e52}}
var e51 = astId{up:&e50,ctx:&e42,id:"="}
var e52 = astCall{up:&e50,ctx:&e42,procParam:&e53} // all this for []
var e53 = astTuple{up:&e52,ctx:&e42,members:[]astNode{&e54,&e55}}
var e54 = astId{up:&e53,ctx:&e42,id:"x"}
var e55 = astCall{up:&e53,ctx:&e42,procParam:&e56}
var e56 = astTuple{up:&e55,ctx:&e42,members:[]astNode{&e57,&e58}}
var e57 = astId{up:&e56,ctx:&e42,"[]"}
var e58 = astTuple{up:&e56,ctx:&e42,members:[]astNode{}} // 0-tuple -> []
var e49 = astId{up:&e47,ctx:&e42,id:"y"}
var e43 = astClosure{up:&e41,ctx:&e13,expr:&e59} // 2nd case
var e59 = astCall{up:&e43,ctx:&e43,procParam:&e60}
var e60 = astTuple{up:&e59,ctx:&e43,members:[]astNode{&e61,&e62}}
var e61 = astId{up:&e60,ctx:&e43,id:";"}
var e62 = astTuple{up:&e60,ctx:&e43,members:[]astNode{&e63,&e74}} // 2 stmts 2nd case
var e63 = astCall{up:&e62,ctx:&e43,procParam:&e64}  // x=`hdx..
var e64 = astTuple{up:&e62,ctx:&e43,members:[]astNode{&e65,&e66}}
var e65 = astId{up:&e64,ctx:&e43,id:"="}
var e66 = astTuple{up:&e64,ctx:&e43,members:[]astNode{&e67,&e68}}
var e67 = astId{up:&e66,ctx:&e43,id:"x"}
var e68 = astCall{up:&e66,ctx:&e43,procParam:&e69} // `hdx +> `tlx
var e69 = astTuple{up:&e68,ctx:&e43,members:[]astNode{&e70,&e71}}
var e70 = astId{up:&e69,ctx:&e43,id:"+>"}
var e71 = astTuple{up:&e69,ctx:&e43,members:[]astNode{&e72,&e73}}
var e72 = astNewId{up:&e71,ctx:&e43,id:"hdx"}
var e73 = astNewId{up:&e71,ctx:&e43,id:"tlx"}
var e74 = astCall{up:&e62,ctx:&e43,procParam:&e75} // 2nd stmt
var e75 = astTuple{up:&e74,ctx:&e43,members:[]astNode{&e76,&e77}}
var e76 = astId{up:&e75,ctx:&e43,id:"+>"}
var e77 = astTuple{up:&e75,ctx:&e43,members:[]astNode{&e78,&e79}}
var e78 = astId{up:&e77,ctx:&e43,id:"hdx"}
var e79 = astCall{up:&e77,ctx:&e43,procParam:&e80}
var e80 = astTuple{up:&e79,ctx:&e43,members:[]astNode{&e81,&e82}}
var e81 = astId{up:&e80,ctx:&e43,id:"append"}
var e82 = astTuple{up:&e80,ctx:&e43,members:[]astNode{&e83,&e84}}
var e83 = astId{up:&e82,ctx:&e43,id:"tlx"}
var e84 = astId{up:&e82,ctx:&e43,id:"y"}

var e6 astCall //print( append([1 2],[3 4]))
var e7 astCall //[1 2 3 4] = append([1 2],print(`a))
var e8 astCall //[1 2 3 4] = append(print(`b),[3 4])
