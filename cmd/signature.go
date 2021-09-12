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
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
				get(pkgName, fset, file, args...)
			}
		}
	},
}

func get(pkgName string, fset *token.FileSet, f *ast.File, tasks ...string) {
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(".", fset, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}
	for _, taskDef := range tasks {
		task := pkg.Scope().Lookup(taskDef)
		fn, _, _ := types.LookupFieldOrMethod(task.Type(), true, pkg, "Do")
		switch t := fn.(type) {
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
				log.Info(typ.Results().At(i).Type())
			}
			tmpl, err := template.ParseFiles("signature.tmpl")
			if err != nil {
				panic(err)
			}
			err = tmpl.Execute(os.Stdout, Task{
				Module: pkgName,
				Name:   taskDef,
				Params: vars,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(signatureCmd)
	signatureCmd.Flags().StringP("pkg", "p", "", "the package your task resides")
	signatureCmd.Flags().StringP("file", "f", "", "the file your task resides")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signatureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signatureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
