/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

	"github.com/iancoleman/strcase"

	log "github.com/sirupsen/logrus"
)

type Var struct {
	Name string
	Type string
}
type Task struct {
	Module string
	Name   string
	Params []Var
}

// signatureCmd represents the signature command
var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "generate a function to New Signature of task",
	Long: `generate a function to New Signature of task. For example:
Your Task definition is

TaskAdd(ctx context.Context, a, b int) 

will generate

func NewTaskAddSignature(a int, b int) *tasks.Signature {
        args := []tasks.Arg{
                {Type: "int", Value: a},
                {Type: "int", Value: b},
        }
        return &tasks.Signature{
                Name: "TaskAdd",
                Args: args,
        }
}
`,
	Run: func(cmd *cobra.Command, args []string) {
		targetPkg, err := cmd.Flags().GetString("pkg")
		if err != nil {
			panic(err)
		}

		targetFile, err := cmd.Flags().GetString("file")
		if err != nil {
			panic(err)
		}
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = filepath.Join(dir, filepath.Dir(targetFile))

		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			panic(err)
		}
		for pkgName, pkg := range pkgs {
			if pkgName != targetPkg {
				log.Info("skip pkg: ", pkgName)
				continue
			}
			for filename, file := range pkg.Files {
				if filepath.Base(filename) != filepath.Base(targetFile) {
					log.Info("skip file", filename)
					continue
				}
				GenerateTaskSignature(dir, pkgName, fset, file, args...)
			}
		}
	},
}

func GenerateTaskSignature(path, pkgName string, fset *token.FileSet, f *ast.File, tasks ...string) {
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(".", fset, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}
	for _, taskDef := range tasks {
		taskObj := pkg.Scope().Lookup(taskDef)
		_, ok := taskObj.(*types.Func)
		if !ok {
			taskObj, _, _ = types.LookupFieldOrMethod(taskObj.Type(), true, pkg, "Do")
		}
		var task Task
		switch t := taskObj.(type) {
		case *types.Func:
			typ := t.Type().(*types.Signature)

			vars := []Var{}
			for i := 1; i < typ.Params().Len(); i++ {
				param := typ.Params().At(i)
				vars = append(vars, Var{Name: param.Name(), Type: param.Type().String()})
			}
			if firstParam := typ.Params().At(0); firstParam.Type().String() != "context.Context" { //TODO: check by using types.Type
				break
			}
			for i := 0; i < typ.Results().Len(); i++ {
				log.Debug(typ.Results().At(i).Type())
			}
			task = Task{
				Module: pkgName,
				Name:   taskDef,
				Params: vars,
			}
		}
		var code strings.Builder
		err = WriteTaskSignature(&code, task)
		if err != nil {
			panic(err)
		}
		f, err := os.OpenFile(filepath.Join(path, strcase.ToSnake(taskDef)+"_signature.go"), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		_, err = f.WriteString(code.String())
		if err != nil {
			panic(err)
		}
	}
}

func WriteTaskSignature(writer io.Writer, task Task) error {
	tmpl, err := template.ParseFS(TmplFiles, "template/signature.tmpl")
	if err != nil {
		return err
	}
	err = tmpl.Execute(writer, task)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(signatureCmd)
	signatureCmd.Flags().StringP("pkg", "p", "", "the package your task resides")
	signatureCmd.Flags().StringP("file", "f", "", "the file your task resides")
	signatureCmd.Flags().StringP("out", "o", "", "the output target")
}
