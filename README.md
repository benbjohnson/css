css [![Build Status](https://drone.io/github.com/benbjohnson/css/status.png)](https://drone.io/github.com/benbjohnson/css/latest) [![Coverage Status](https://coveralls.io/repos/benbjohnson/css/badge.png?branch=master)](https://coveralls.io/r/benbjohnson/css?branch=master) [![GoDoc](https://godoc.org/github.com/benbjohnson/css?status.png)](https://godoc.org/github.com/benbjohnson/css) ![Project status](http://img.shields.io/status/alpha.png?color=red)
===

This package provides a CSS parser and scanner in pure Go. It is an
implementation as specified in the W3C's [CSS Syntax Module Level 3](css3-syntax).

For documentation on how to use this package, please see the [godoc][godoc].

[css3-syntax]: http://www.w3.org/TR/css3-syntax/
[godoc]: https://godoc.org/github.com/benbjohnson/css


## Project Status

The scanner and parser are fully compliant with the CSS3 specification.
The printer will print nodes generated from the scanner and parser, however,
it is not fully compliant with the [CSS3 serialization][serialization] spec.
Additionally, the printer does not provide an option to collapse whitespace
although that will be added in the future.

This project has 100% test coverage, however, it is still a new project.
Please report any bugs you experience or let me know where the documentation
can be clearer.

[serialization]: http://www.w3.org/TR/css3-syntax/#serialization


## Caveats

The CSS scanner in this package only supports UTF-8 encoding. The @charset
directive will be ignored. If you need to scan a different encoding then
please convert it to UTF-8 first using a tool such as [iconv][iconv].

[iconv]: http://en.wikipedia.org/wiki/Iconv
