package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		panic(err)
	}
	conf := types.Config{Importer: importer.Default()}
	var info types.Info
	pkg, err := conf.Check("tasks", fset, []*ast.File{f}, &info)
	if err != nil {
		panic(err)
	}
	log.Info(info)
	task := pkg.Scope().Lookup("TaskAdd")
	fn, _, _ := types.LookupFieldOrMethod(task.Type(), true, pkg, "Do")
	switch t := fn.(type) {
	case *types.Func:
		typ := t.Type().(*types.Signature)
		for i := 0; i < typ.Params().Len(); i++ {
			fmt.Println(typ.Params().At(i))
		}
		if firstParam := typ.Params().At(0); firstParam.Type().String() != "context.Context" { //TODO: check by using types.Type
			break
		}

		for i := 0; i < typ.Results().Len(); i++ {
			fmt.Println(typ.Results().At(i))
		}
	}
}
