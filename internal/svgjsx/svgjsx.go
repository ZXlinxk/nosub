package svgjsx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Options configures the SVG to JSX conversion.
type Options struct {
	Name         string
	TypeScript   bool
	Tailwind     bool
	CurrentColor bool
}

const (
	svgNS   = "http://www.w3.org/2000/svg"
	xhtmlNS = "http://www.w3.org/1999/xhtml"
	xlinkNS = "http://www.w3.org/1999/xlink"
	xmlNS   = "http://www.w3.org/XML/1998/namespace"
)

type attr struct{ name, value string }

type node struct {
	tag      string
	attrs    []attr
	children []*node
	text     string
}

func (n *node) isText() bool { return n.tag == "" }

// Convert transforms raw SVG bytes into a React JSX/TSX component.
func Convert(svg []byte, o Options) (string, error) {
	if o.Name == "" {
		o.Name = "Icon"
	}

	root, err := parseSVG(svg)
	if err != nil {
		return "", err
	}

	prepare(root, true, o)

	var b strings.Builder
	if o.TypeScript {
		b.WriteString("import type { SVGProps } from \"react\";\n\n")
		fmt.Fprintf(&b, "export default function %s(props: SVGProps<SVGSVGElement>) {\n", o.Name)
	} else {
		fmt.Fprintf(&b, "export default function %s(props) {\n", o.Name)
	}
	b.WriteString("  return (\n")
	renderJSX(&b, root, 4, true)
	b.WriteString("\n  );\n}\n")

	return b.String(), nil
}

func parseSVG(svg []byte) (*node, error) {
	dec := xml.NewDecoder(bytes.NewReader(svg))
	var stack []*node
	var root *node

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if root == nil && t.Name.Local != "svg" {
				return nil, errors.New("root element must be svg")
			}
			if !keepElement(t.Name) {
				if err := dec.Skip(); err != nil {
					return nil, err
				}
				continue
			}
			n := &node{tag: t.Name.Local, attrs: keepAttrs(t.Attr)}
			if root == nil {
				root = n
			} else {
				top := stack[len(stack)-1]
				top.children = append(top.children, n)
			}
			stack = append(stack, n)

		case xml.EndElement:
			stack = stack[:len(stack)-1]

		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			top := stack[len(stack)-1]
			if top.tag == "style" || strings.TrimSpace(string(t)) != "" {
				top.children = append(top.children, &node{text: string(t)})
			}
		}
	}

	if root == nil {
		return nil, errors.New("empty svg")
	}
	return root, nil
}

// keepElement drops everything editors add: their tags live in foreign namespaces,
// which React could not render anyway.
func keepElement(name xml.Name) bool {
	switch name.Space {
	case "", svgNS, xhtmlNS:
		return name.Local != "metadata" && name.Local != "script"
	}
	return false
}

func keepAttrs(in []xml.Attr) []attr {
	var out []attr
	for _, a := range in {
		local := a.Name.Local
		switch a.Name.Space {
		case "":
			// Inline handlers are strings here and React throws on those.
			if !strings.HasPrefix(strings.ToLower(local), "on") {
				out = append(out, attr{local, a.Value})
			}
		case xlinkNS, "xlink":
			out = append(out, attr{"xlink:" + local, a.Value})
		case xmlNS, "xml":
			out = append(out, attr{"xml:" + local, a.Value})
		}
	}
	return out
}

func prepare(n *node, root bool, o Options) {
	if n.isText() {
		return
	}
	var decls []styleDecl
	n.attrs = slices.DeleteFunc(n.attrs, func(a attr) bool {
		if a.name == "style" {
			decls = parseStyle(a.value)
			return true
		}
		return false
	})

	if o.CurrentColor {
		for i, a := range n.attrs {
			if (a.name == "fill" || a.name == "stroke") && canCurrentColor(a.value) {
				n.attrs[i].value = "currentColor"
			}
		}
		for i, d := range decls {
			if (d.key == "fill" || d.key == "stroke") && canCurrentColor(d.val) {
				decls[i].val = "currentColor"
			}
		}
	}
	if o.Tailwind {
		decls = applyTailwind(n, decls, root)
	}
	for i, a := range n.attrs {
		n.attrs[i].name = reactName(a.name)
	}
	if len(decls) > 0 {
		n.attrs = append(n.attrs, attr{"style", formatStyleObject(decls)})
	}
	for _, c := range n.children {
		prepare(c, false, o)
	}
}

// applyTailwind moves presentation attributes and declarations into className
// and returns the declarations with no utility class.
func applyTailwind(n *node, decls []styleDecl, root bool) []styleDecl {
	take := func(name string) string {
		var val string
		n.attrs = slices.DeleteFunc(n.attrs, func(a attr) bool {
			if a.name == name {
				val = a.value
				return true
			}
			return false
		})
		return val
	}
	// Declarations win over attributes, as they do in CSS.
	pull := func(name string) string {
		val := take(name)
		decls = slices.DeleteFunc(decls, func(d styleDecl) bool {
			if d.key == name {
				val = d.val
				return true
			}
			return false
		})
		return val
	}

	var classes []string
	if c := take("class"); c != "" {
		classes = append(classes, c)
	}
	if root {
		sizes, w, h := tailwindSize(take("width"), take("height"))
		classes = append(classes, sizes...)
		if w != "" {
			n.attrs = append(n.attrs, attr{"width", w})
		}
		if h != "" {
			n.attrs = append(n.attrs, attr{"height", h})
		}
	}
	for _, name := range []string{"fill", "stroke"} {
		if v := pull(name); v != "" {
			if cls, kept := tailwindColor(name, v); kept {
				n.attrs = append(n.attrs, attr{name, v})
			} else {
				classes = append(classes, cls)
			}
		}
	}
	if v := pull("stroke-width"); v != "" {
		classes = append(classes, tailwindStrokeWidth(v))
	}
	if v := pull("opacity"); v != "" {
		classes = append(classes, tailwindOpacity(v))
	}
	if len(classes) > 0 {
		n.attrs = append([]attr{{"class", strings.Join(classes, " ")}}, n.attrs...)
	}
	return decls
}

func canCurrentColor(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v != "none" && v != "transparent" && !strings.HasPrefix(v, "url(")
}

func reactName(name string) string {
	switch {
	case name == "class":
		return "className"
	case strings.HasPrefix(name, "data-"), strings.HasPrefix(name, "aria-"):
		return name
	}
	return toCamelCase(name)
}

func renderJSX(b *strings.Builder, n *node, indent int, isRoot bool) {
	pad := strings.Repeat(" ", indent)
	b.WriteString(pad)
	b.WriteByte('<')
	b.WriteString(n.tag)

	for _, a := range n.attrs {
		b.WriteByte(' ')
		if a.name == "style" {
			b.WriteString(a.value)
			continue
		}
		b.WriteString(a.name)
		b.WriteString("=\"")
		b.WriteString(escapeAttr(a.value))
		b.WriteByte('"')
	}

	if isRoot {
		b.WriteString(" {...props}")
	}

	if len(n.children) == 0 {
		b.WriteString(" />")
		return
	}

	b.WriteByte('>')

	if n.tag == "style" {
		var css strings.Builder
		for _, c := range n.children {
			css.WriteString(c.text)
		}
		raw := css.String()
		raw = strings.ReplaceAll(raw, "`", "\\`")
		raw = strings.ReplaceAll(raw, "${", "\\${")
		b.WriteString("{`")
		b.WriteString(raw)
		b.WriteString("`}</style>")
		return
	}

	if len(n.children) == 1 && n.children[0].isText() {
		b.WriteString(escapeJSXText(n.children[0].text))
		b.WriteString("</")
		b.WriteString(n.tag)
		b.WriteByte('>')
		return
	}

	b.WriteByte('\n')
	for _, c := range n.children {
		if c.isText() {
			b.WriteString(strings.Repeat(" ", indent+2))
			b.WriteString(escapeJSXText(strings.TrimSpace(c.text)))
		} else {
			renderJSX(b, c, indent+2, false)
		}
		b.WriteByte('\n')
	}

	b.WriteString(pad)
	b.WriteString("</")
	b.WriteString(n.tag)
	b.WriteByte('>')
}

func escapeAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", `"`, "&quot;").Replace(s)
}

func escapeJSXText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '{':
			b.WriteString(`{"{"}`)
		case '}':
			b.WriteString(`{"}"}`)
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			b.WriteString("&amp;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
