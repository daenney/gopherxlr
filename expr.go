package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"code.dny.dev/gopherxlr/dbus"
	"code.dny.dev/gopherxlr/websocket"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
)

type Env struct {
	StatusChange websocket.StatusChange `expr:"status"`
	Context      context.Context        `expr:"ctx"`
}

func (Env) PlayPause(ctx context.Context, name string) error {
	return dbus.ToggleMediaPlayback(ctx, name)
}

func (Env) RunCommand(ctx context.Context, name string, args ...string) error {
	_, err := RunCommand(ctx, name, args...)
	if err != nil {
		return err
	}
	return nil
}

type Program struct {
	file string
	prog *vm.Program
}

func LoadPrograms(dir string) ([]Program, error) {
	res := []Program{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			prog, err := expr.Compile(string(src), expr.Env(Env{}), expr.Patch(patcher{}))
			if err != nil {
				return err
			}
			res = append(res, Program{file: path, prog: prog})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

type patcher struct{}

// Until https://github.com/expr-lang/expr/pull/602 is merged
func (patcher) Visit(node *ast.Node) {
	switch call := (*node).(type) {
	case *ast.CallNode:
		fn := call.Callee.Type()
		if fn == nil {
			return
		}
		if fn.Kind() != reflect.Func {
			return
		}
		switch fn.NumIn() {
		case 0:
			return
		case 1:
			if fn.In(0).String() != "context.Context" {
				return
			}
		default:
			if fn.In(0).String() != "context.Context" &&
				fn.In(1).String() != "context.Context" {
				return
			}
		}
		ast.Patch(node, &ast.CallNode{
			Callee: call.Callee,
			Arguments: append([]ast.Node{
				&ast.IdentifierNode{Value: "ctx"},
			}, call.Arguments...),
		})
	}
}
