/*
The program we hand convert to AST and interpret is:

    `append = {
        $ = (`x,`y); # 2 input lists
        caseP [
            { x=[]; y }
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
    up *astNode
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
