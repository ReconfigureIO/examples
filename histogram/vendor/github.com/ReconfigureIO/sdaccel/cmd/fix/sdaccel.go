// Copyright 2017 Reconfigure.io.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"go/ast"
)

func init() {
	register(sdaccel)
}

var sdaccel = fix{
	name: "sdaccel",
	date: "2017-12-12",
	f:    sdaccelFix,
	desc: `Change imports of sdaccel to github.com/ReconfigureIO/sdaccel`,
}

func sdaccelFix(f *ast.File) bool {
	ret := false
	ret = rewriteImport(f, "xcl", "github.com/ReconfigureIO/sdaccel/xcl") || ret
	ret = rewriteImport(f, "sdaccel", "github.com/ReconfigureIO/sdaccel") || ret
	ret = rewriteImport(f, "sdaccel/control", "github.com/ReconfigureIO/sdaccel/control") || ret
	ret = rewriteImport(f, "axi", "github.com/ReconfigureIO/sdaccel/axi") || ret
	ret = rewriteImport(f, "axi/protocol", "github.com/ReconfigureIO/sdaccel/axi/protocol") || ret
	ret = rewriteImport(f, "axi/arbitrate", "github.com/ReconfigureIO/sdaccel/axi/arbitrate") || ret
	ret = rewriteImport(f, "axi/memory", "github.com/ReconfigureIO/sdaccel/axi/memory") || ret
	return ret
}
