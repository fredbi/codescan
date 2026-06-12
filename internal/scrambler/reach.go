// SPDX-FileCopyrightText: Copyright 2015-2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package scrambler

import (
	"go/ast"
	"go/token"
	"go/types"
	"maps"

	"golang.org/x/tools/go/packages"
)

// declSite locates an object's declaration AST for the reachability walk.
type declSite struct {
	node ast.Node // *ast.FuncDecl | *ast.TypeSpec | *ast.ValueSpec
	pkg  *packages.Package
	file *ast.File
}

// closure computes and holds the set of package-level objects to KEEP: the
// transitive closure of the annotation roots over the type graph, per the
// reachability rules in .claude/plans/scrambler-L0-build.md (R1–R6).
type closure struct {
	keep map[types.Object]bool

	objToSite     map[types.Object]declSite
	uses          map[*ast.Ident]types.Object // merged TypesInfo.Uses over in-scope pkgs
	inScopePkg    map[*types.Package]bool
	methodsOf     map[*types.TypeName][]types.Object // R3: type → its methods
	valuesOf      map[*types.TypeName][]types.Object // R4: type → package-level consts/vars of that type
	constSiblings map[types.Object][]types.Object    // const-block grouping (iota + ref safety)
}

func computeClosure(l *loaded, roots map[types.Object]bool) *closure {
	c := &closure{
		keep:          map[types.Object]bool{},
		objToSite:     map[types.Object]declSite{},
		uses:          map[*ast.Ident]types.Object{},
		inScopePkg:    map[*types.Package]bool{},
		methodsOf:     map[*types.TypeName][]types.Object{},
		valuesOf:      map[*types.TypeName][]types.Object{},
		constSiblings: map[types.Object][]types.Object{},
	}

	for _, pkg := range l.pkgs {
		c.inScopePkg[pkg.Types] = true
		maps.Copy(c.uses, pkg.TypesInfo.Uses)
	}
	c.index(l)

	work := make([]types.Object, 0, len(roots))
	for r := range roots {
		if !c.keep[r] {
			c.keep[r] = true
			work = append(work, r)
		}
	}
	for len(work) > 0 {
		obj := work[len(work)-1]
		work = work[:len(work)-1]
		work = c.expand(obj, work)
	}
	return c
}

// index builds the decl-site, method, value and const-sibling maps from the
// in-scope files.
func (c *closure) index(l *loaded) {
	for _, sf := range l.files {
		for _, decl := range sf.ast.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				obj := sf.pkg.TypesInfo.Defs[d.Name]
				if obj == nil {
					continue
				}
				c.objToSite[obj] = declSite{node: d, pkg: sf.pkg, file: sf.ast}
				if tn := methodRecvTypeName(obj); tn != nil {
					c.methodsOf[tn] = append(c.methodsOf[tn], obj)
				}
			case *ast.GenDecl:
				c.indexGenDecl(sf, d)
			}
		}
	}
}

func (c *closure) indexGenDecl(sf *sourceFile, d *ast.GenDecl) {
	pkg := sf.pkg
	var constGroup []types.Object
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if obj := pkg.TypesInfo.Defs[s.Name]; obj != nil {
				c.objToSite[obj] = declSite{node: s, pkg: pkg, file: sf.ast}
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				obj := pkg.TypesInfo.Defs[name]
				if obj == nil {
					continue
				}
				c.objToSite[obj] = declSite{node: s, pkg: pkg, file: sf.ast}
				if tn := namedTypeNameOf(obj.Type()); tn != nil && c.inScopePkg[tn.Pkg()] {
					c.valuesOf[tn] = append(c.valuesOf[tn], obj)
				}
				if d.Tok == token.CONST {
					constGroup = append(constGroup, obj)
				}
			}
		}
	}
	// Group const-block members: keeping any one keeps all, so iota sequences
	// stay intact and no kept member references a dropped sibling.
	for _, o := range constGroup {
		c.constSiblings[o] = constGroup
	}
}

// expand keeps everything reachable from obj per the rules, enqueueing newcomers.
func (c *closure) expand(obj types.Object, work []types.Object) []types.Object {
	if site, ok := c.objToSite[obj]; ok {
		work = c.walkRefs(site, work) // R1/R2
	}
	if tn, ok := obj.(*types.TypeName); ok {
		for _, m := range c.methodsOf[tn] { // R3
			work = c.add(m, work)
		}
		for _, v := range c.valuesOf[tn] { // R4
			work = c.add(v, work)
		}
	}
	for _, sib := range c.constSiblings[obj] { // const-block grouping
		work = c.add(sib, work)
	}
	return work
}

// walkRefs adds every in-scope package-level object referenced by the site. For
// functions it walks only the signature (receiver + FuncType), never the body —
// the body is emptied, so body-only references must not keep types alive.
func (c *closure) walkRefs(site declSite, work []types.Object) []types.Object {
	visit := func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if used := c.uses[id]; c.isInScopePkgLevel(used) {
				work = c.add(used, work)
			}
		}
		return true
	}
	switch d := site.node.(type) {
	case *ast.FuncDecl:
		// A body carrying an annotation (e.g. a nested swagger:response type) is
		// kept verbatim, so walk the whole decl to reach the types it uses.
		// Otherwise the body is emptied, so walk only the signature.
		if bodyHasSwaggerAnnotation(d, site.file.Comments) {
			ast.Inspect(d, visit)
			break
		}
		if d.Recv != nil {
			ast.Inspect(d.Recv, visit)
		}
		if d.Type != nil {
			ast.Inspect(d.Type, visit)
		}
	default:
		ast.Inspect(site.node, visit)
	}
	return work
}

func (c *closure) add(obj types.Object, work []types.Object) []types.Object {
	if obj == nil || c.keep[obj] {
		return work
	}
	c.keep[obj] = true
	return append(work, obj)
}

func (c *closure) isInScopePkgLevel(obj types.Object) bool {
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if !c.inScopePkg[obj.Pkg()] {
		return false
	}
	return obj.Parent() == obj.Pkg().Scope()
}

// methodRecvTypeName returns the named type a method is defined on, or nil.
func methodRecvTypeName(obj types.Object) *types.TypeName {
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return nil
	}
	return namedTypeNameOf(sig.Recv().Type())
}

// namedTypeNameOf unwraps a pointer and returns the underlying *types.Named's
// TypeName, or nil for unnamed/basic types.
func namedTypeNameOf(t types.Type) *types.TypeName {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := t.(*types.Named); ok {
		return n.Obj()
	}
	return nil
}
