css
===

This package provides a CSS parser and scanner in pure Go. It is an
implementation as specified in the W3C's [CSS Syntax Module Level 3](css3-syntax).

[css3-syntax]: http://www.w3.org/TR/css3-syntax/


## Caveats

The CSS scanner in this package only supports UTF-8 encoding. The @charset
directive will be ignored. If you need to scan a different encoding then
please convert it to UTF-8 first using a tool such as [iconv][iconv].

[iconv]: http://en.wikipedia.org/wiki/Iconv
