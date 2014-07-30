css
===

This package provides a CSS parser and scanner in pure Go. It is an
implementation as specified in the W3C's [CSS Syntax Module Level 3](css3-syntax).

[css3-syntax]: http://www.w3.org/TR/css3-syntax/


## Getting Started

### Parsing a stylesheet

The `css.Parser` is the most useful type for most users. To use it, simply
provide a reader that contains your stylesheet text:

```go
// Open your CSS file.
f, err := os.Open("mystyle.css")
if err != nil {
	log.Fatal(err)
}
defer f.Close()

// Instantiate a parser and parse the stylesheet.
p := css.NewParser(f)
ss, err := p.Parse()
if err != nil {
	log.Fatal("css parse error: ", err)
}

// Do something with your stylesheet.
...
```


### Traversing the stylesheet

You can iterate over your stylesheet by implementing `css.Visitor`:

```
// RulePrinter prints every rule name in a stylesheet.
type RulePrinter struct {}

// Visit is called for every node in the stylesheet.
func (v *RulePrinter) Visit(node css.Node) css.Visitor {
	switch node := node.(type) {
	case *css.QualifiedRule:
		fmt.Println(node.Name)
	}
	return v
}
```

