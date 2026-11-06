package svgjsx

import (
	"strings"
	"testing"
)

const sample = `<?xml version="1.0" encoding="UTF-8"?>
<!-- Generator: Sketch 96 -->
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
     xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape"
     xmlns:sodipodi="http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd"
     width="24" height="24" viewBox="0 0 24 24" fill="none" class="icon" inkscape:version="1.2">
  <metadata>junk</metadata>
  <sodipodi:namedview id="base"/>
  <title>Logo</title>
  <path d="M0 0L24 24" stroke="#000000" stroke-width="2" stroke-linecap="round" style="opacity:.5;fill-rule:evenodd" sodipodi:nodetypes="cc"/>
  <use xlink:href="#a" clip-path="url(#c)" fill="#3b82f6"/>
</svg>`

func has(t *testing.T, s string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(s, w) {
			t.Errorf("missing %q in:\n%s", w, s)
		}
	}
}

func hasNot(t *testing.T, s string, bad ...string) {
	t.Helper()
	for _, w := range bad {
		if strings.Contains(s, w) {
			t.Errorf("should not contain %q:\n%s", w, s)
		}
	}
}

func TestPlainJSX(t *testing.T) {
	out, err := Convert([]byte(sample), Options{Name: "Logo"})
	if err != nil {
		t.Fatal(err)
	}
	has(t, out,
		"export default function Logo(props) {",
		"strokeWidth=\"2\"", "strokeLinecap=\"round\"", "clipPath=\"url(#c)\"", "xlinkHref=\"#a\"",
		"className=\"icon\"", "viewBox=\"0 0 24 24\"",
		`style={{ opacity: ".5", fillRule: "evenodd" }}`,
		"{...props}", "<title>Logo</title>", "<path", "/>")
	hasNot(t, out, "inkscape", "sodipodi", "metadata", "Generator", "xmlns:", "<?xml", "class=", "stroke-width", "junk")
}

func TestSpreadPropsComesLast(t *testing.T) {
	out, _ := Convert([]byte(`<svg viewBox="0 0 1 1" fill="red"><g/></svg>`), Options{Name: "A"})
	has(t, out, `fill="red" {...props}>`)
}

func TestTypeScript(t *testing.T) {
	out, err := Convert([]byte(sample), Options{Name: "Logo", TypeScript: true})
	if err != nil {
		t.Fatal(err)
	}
	has(t, out, `import type { SVGProps } from "react";`, "export default function Logo(props: SVGProps<SVGSVGElement>) {")
}

func TestTailwind(t *testing.T) {
	out, err := Convert([]byte(sample), Options{Name: "Logo", Tailwind: true})
	if err != nil {
		t.Fatal(err)
	}
	has(t, out,
		`className="icon size-6 fill-none"`,
		`className="stroke-black stroke-2 opacity-50"`,
		`className="fill-blue-500"`,
		`strokeLinecap="round"`,
		`style={{ fillRule: "evenodd" }}`,
		`viewBox="0 0 24 24"`)
	hasNot(t, out, `width="24"`, `height="24"`, `fill="none"`, `stroke="#000000"`, `strokeWidth="2"`)
}

func TestTailwindValues(t *testing.T) {
	cases := []struct{ in, want string }{
		{`<svg width="16" height="16"/>`, `size-4`},
		{`<svg width="18" height="18"/>`, `size-[18px]`},
		{`<svg width="24" height="32"/>`, `w-6 h-8`},
		{`<svg width="30" height="32"/>`, `w-[30px] h-8`},
		{`<svg fill="#123456"/>`, `fill-[#123456]`},
		{`<svg fill="#FFF"/>`, `fill-white`},
		{`<svg fill="currentColor"/>`, `fill-current`},
		{`<svg stroke="#ef4444"/>`, `stroke-red-500`},
		{`<svg stroke-width="1.5"/>`, `stroke-[1.5]`},
		{`<svg opacity="0.33"/>`, `opacity-[.33]`},
		{`<svg fill="url(#g)"/>`, `fill="url(#g)"`},
	}
	for _, c := range cases {
		out, err := Convert([]byte(c.in), Options{Name: "A", Tailwind: true})
		if err != nil {
			t.Fatal(err)
		}
		has(t, out, c.want)
	}
}

func TestCurrentColor(t *testing.T) {
	out, _ := Convert([]byte(`<svg fill="none"><path fill="#123" stroke="red"/><path fill="url(#a)"/></svg>`), Options{Name: "A", CurrentColor: true})
	has(t, out, `fill="none"`, `fill="currentColor"`, `stroke="currentColor"`, `fill="url(#a)"`)
	hasNot(t, out, "#123", "red")
}

func TestStyleElementAndText(t *testing.T) {
	out, err := Convert([]byte(`<svg><style>.a{fill:red}</style><text x="1">a { b } &lt; c</text></svg>`), Options{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	has(t, out, "<style>{`.a{fill:red}`}</style>", `a {"{"} b {"}"} &lt; c`)
}

func TestForeignNamespacesAndHandlersRemoved(t *testing.T) {
	in := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:i="http://ns.adobe.com/AdobeIllustrator/10.0/" i:viewOrigin="0 0" onload="x()"><i:pgf>data</i:pgf><script>alert(1)</script><rect/></svg>`
	out, err := Convert([]byte(in), Options{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	has(t, out, "<rect />")
	hasNot(t, out, "pgf", "viewOrigin", "onload", "script", "data")
}

func TestStyleBeatsAttribute(t *testing.T) {
	out, _ := Convert([]byte(`<svg><path fill="#000" style="fill:#fff"/></svg>`), Options{Name: "A", Tailwind: true})
	has(t, out, `className="fill-white"`)
}

func TestArbitraryValueHasNoSpaces(t *testing.T) {
	out, _ := Convert([]byte(`<svg fill="rgb(1, 2, 3)"/>`), Options{Name: "A", Tailwind: true})
	has(t, out, `fill-[rgb(1,_2,_3)]`)
}

func TestDataAndAriaKept(t *testing.T) {
	out, _ := Convert([]byte(`<svg aria-hidden="true" data-foo="1" xml:space="preserve"/>`), Options{Name: "A"})
	has(t, out, `aria-hidden="true"`, `data-foo="1"`, `xmlSpace="preserve"`)
}

func TestInvalid(t *testing.T) {
	if _, err := Convert([]byte(`<svg><path></svg>`), Options{Name: "A"}); err == nil {
		t.Fatal("want error on malformed xml")
	}
	if _, err := Convert([]byte(`<div/>`), Options{Name: "A"}); err == nil {
		t.Fatal("want error when root is not svg")
	}
}

func TestComponentName(t *testing.T) {
	cases := map[string]string{
		"arrow-left.svg":        "ArrowLeft",
		"icons/user_circle.svg": "UserCircle",
		"24 px icon.svg":        "Icon24PxIcon",
		"a.b.svg":               "AB",
		"---.svg":               "Icon",
	}
	for in, want := range cases {
		if got := ComponentName(in); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}
