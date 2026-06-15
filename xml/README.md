# xml

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/xml.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/xml)

Builds XML markup with the same component model as the [`html`](../html) and
[`svg`](../svg) packages — but because XML has no fixed vocabulary, elements and
attributes are created by **generic** constructors rather than one function per
tag. On top of those it adds the constructs that are specific to XML documents:
comments, CDATA sections, processing instructions, the XML declaration and
document type declarations.

```go
import "github.com/ungerik/go-mx/xml"

doc := xml.NewDocument(
    xml.Element("note",
        xml.Attrib("id", 42),
        xml.Element("to", "Tove"),
        xml.Element("from", "Jani"),
        xml.Comment("the message body"),
        xml.Element("body", xml.CDATA("unescaped <raw> & text")),
    ),
)
// <?xml version="1.0" encoding="UTF-8"?>
// <note id="42">
//   <to>Tove</to>
//   <from>Jani</from>
//   <!-- the message body -->
//   <body>
//     <![CDATA[unescaped <raw> & text]]>
//   </body>
// </note>
```

## Relationship to the `mx` package

`xml` is a thin, XML-flavored vocabulary over the root
[`mx`](https://pkg.go.dev/github.com/ungerik/go-mx) package, which owns the core
abstractions (`Component`, `Element`, `Attrib`, `Writer`). The `mx.Writer`
already knows how to emit comments and CDATA sections, and its `CheckedWriter`
escapes text and attribute values, balances tags and validates CDATA — this
package names the XML-level constructs that build on it.

XML and HTML escape the same five characters (`&`, `<`, `>`, `"`, `'`), so the
default `CheckedWriter` is used unchanged; output uses double-quoted attribute
values, the XML convention.

## Elements

XML tag names are arbitrary, so there are no per-tag constructors — every
element is built with the generic functions:

| Constructor                         | Renders                              |
|-------------------------------------|--------------------------------------|
| `Element(name, attribsChildren...)` | `<name ...>children</name>`          |
| `EmptyElement(name, attribs...)`    | `<name .../>` (self-closing)         |
| `ElementNS(prefix, name, ...)`      | `<prefix:name ...>...</prefix:name>` |
| `EmptyElementNS(prefix, name, ...)` | `<prefix:name .../>`                 |

`Element` with no children renders an explicit close tag (`<name></name>`);
`EmptyElement` renders the self-closing form. The `…NS` shortcuts build a
namespace-qualified name `prefix:name` (an empty prefix falls back to the plain
name) — bind the prefix to a URI with an `XMLNSPrefix` attribute:

```go
xml.ElementNS("soap", "Envelope",
    xml.XMLNSPrefix("soap", "http://schemas.xmlsoap.org/soap/envelope/"),
    xml.ElementNS("soap", "Body", /* … */),
)
// <soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body>…</soap:Body></soap:Envelope>
```

### Children

Children are passed as `...any` and converted by `mx`: a `string` becomes
escaped `Text`, a `Component` (including the constructs below) passes through,
and other values are stringified. Use `Text`/`Textf` for explicit text and
`Raw`/`RawBytes` for pre-formed markup that must not be escaped.

## Attributes

| Constructor                         | Renders                  |
|-------------------------------------|--------------------------|
| `Attrib(name, value)`               | `name="value"`           |
| `AttribNS(prefix, name, value)`     | `prefix:name="value"`    |

`Attrib` is generic over [`AttribValue`](attributes.go) (a string or any integer
or floating-point type): strings pass through and floats render as plain
decimals, never scientific notation, so `Attrib("w", 100)`, `Attrib("ratio", 0.5)`
and `Attrib("id", "x1")` all work.

To pass a dynamic set of attributes, build an `Attribs` (an alias for
[`mx.Attribs`](https://pkg.go.dev/github.com/ungerik/go-mx#Attribs), a
`map[string]any` of name to value) and hand it to an element like any other
argument; it renders sorted, with `id` first and `class` second.

The predefined `xml:`/`xmlns` attributes have dedicated constructors because they
mean the same thing in every document:

| Constructor / constant       | Renders                       |
|------------------------------|-------------------------------|
| `XMLNS(uri)`                 | `xmlns="uri"`                 |
| `XMLNSPrefix(prefix, uri)`   | `xmlns:prefix="uri"`          |
| `XMLLang(lang)`              | `xml:lang="lang"`             |
| `XMLBase(uri)`               | `xml:base="uri"`              |
| `XMLID(id)`                  | `xml:id="id"`                 |
| `XMLSpace(value)`            | `xml:space="value"`           |
| `XMLSpacePreserve`           | `xml:space="preserve"`        |
| `XMLSpaceDefault`            | `xml:space="default"`         |

## XML-specific constructs

| Construct                  | Renders                                  |
|----------------------------|------------------------------------------|
| `Comment("text")`          | `<!-- text -->`                          |
| `CDATA("text")`            | `<![CDATA[text]]>`                       |
| `ProcInst{Target, Data}`   | `<?Target Data?>`                        |
| `Declaration`              | `<?xml version="1.0" encoding="UTF-8"?>` |
| `Decl(version, encoding)`  | `<?xml version="..." encoding="..."?>`   |
| `Doctype("text")`          | `<!DOCTYPE text>`                        |

`Comment` and `CDATA` are `string` types (like `Text` and `Raw`), so they read
as values: `xml.Comment("note")`, `xml.CDATA("<raw>")`. Both refuse to emit
malformed markup — a `Comment` containing `--`, or a `CDATA` containing its
terminator `]]>`, fails to render with an error rather than producing invalid
XML. This is the package's [deferred-error pattern](../doc.go): the problem
surfaces when the affected subtree is rendered.

`ProcInst` validates its `Target` (non-empty, not the reserved `xml`, no
whitespace) and that neither part contains `?>`. `Declaration`/`Decl` build the
XML declaration and `Doctype` a document type declaration. They are bare `Raw`
values with no trailing newline; the `mx.CheckedWriter` recognizes the closing
`?>` of a declaration or processing instruction and breaks the line after it, so
each sits on its own line in both compact and indented output (this also applies
to HTML rendering, where `?>` simply never occurs). `Declaration`/`Decl` and
`Doctype` are `Raw`; `ProcInst` is a struct component, but all of them are
`mx.Component`s, so they compose freely or serve as a `Document`'s
`Declaration`/`Prolog`.

## Conditional rendering and iteration

`If`, `Iff`, `ForEach` and `ForEachIter` mirror the `html` and `mx` helpers:

```go
xml.Element("list",
    xml.ForEach(items, func(it Item) *mx.Element {
        return xml.Element("item", xml.Attrib("id", it.ID), it.Name)
    }),
    xml.If(footer != "", xml.Element("footer", footer)),
)
```

## Documents and serving

`Document` assembles a complete document — an optional `Declaration`, optional
`Prolog` (comments, processing instructions, a `Doctype`) and a single `Root`
element. `NewDocument(root)` defaults the declaration to the standard
`Declaration`. The writer breaks the line after the declaration's `?>`, so the
prolog or root starts on the next line in both compact and indented output (an
indenting writer, as `Serve` and `Document.HandleHTTP` use, further indents
nested elements).

```go
doc := &xml.Document{
    Declaration: xml.Declaration,
    Prolog:      mx.Components{xml.Doctype("note"), xml.Comment("generated")},
    Root:        xml.Element("note", /* … */),
}
http.HandleFunc("/feed.xml", doc.HandleHTTP) // Content-Type: application/xml
```

`Serve(addr, component)` serves any component as indented XML — `doc.Serve(addr)`
is the shorthand for a `Document` — and `String(c)` renders a component to a
string for tests and one-off use.

## Escaping helpers

`Escape`, `WriteEscaped` and `WriteRaw` mirror the `html` helpers for the five
XML predefined entities (`&amp; &lt; &gt; &quot; &apos;`).

## Online sources

- [Extensible Markup Language (XML) 1.0 (W3C Recommendation)](https://www.w3.org/TR/xml/)
- [Namespaces in XML 1.0](https://www.w3.org/TR/xml-names/)
- [MDN Web Docs — XML introduction](https://developer.mozilla.org/en-US/docs/Web/XML/Guides/XML_introduction)
