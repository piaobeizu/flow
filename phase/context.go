/*
 @Version : 1.0
 @Author  : steven.wong
 @Email   : 'wangxk1991@gamil.com'
 @Time    : 2023/04/03 16:34:16
 Desc     :
*/

package phase

import (
	"context"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific args and
// parsed command-line options.

type Context struct {
	Context *context.Context
	Data    map[string]any
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(ctx context.Context) *Context {
	return &Context{Context: &ctx, Data: map[string]any{}}
}

// Set sets a context flag to a value.
func (cCtx *Context) Set(name string, value any) {
	cCtx.Data[name] = value
}

// Set sets a context flag to a value.
func (cCtx *Context) Get(name string) any {
	if v, ok := cCtx.Data[name]; ok {
		return v
	}
	return nil
}

// Set sets a context flag to a value.
func (cCtx *Context) GetCtx() *context.Context {
	return cCtx.Context
}
