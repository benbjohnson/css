/*
Package css implements a CSS3 compliant scanner and parser. This is meant to
be a low-level library for extracting a CSS3 abstract syntax tree from raw
CSS text.

This package can be used for building tools to validate, optimize and format
CSS text.


Basics

CSS parsing occurs in two steps. First the scanner breaks up a stream of code
points (runes) into tokens. These tokens represent the most basic units of
the CSS syntax tree such as identifiers, whitespace, and strings. The second
step is to feed these tokens into the parser which creates the abstract syntax
tree (AST) based on the context of the tokens.

Unlike many language parsers, the abstract syntax tree for CSS saves many of the
original tokens in the stream so they can be reparsed at different levels. For
example, parsing a @media query will save off the raw tokens found in the
{-block so they can be reparsed as a full style sheet. This package doesn't
understand the specifics of how to parse different types of at-rules (such as
@media queries) so it defers that to the user to handle parsing.


Abstract Syntax Tree

The CSS3 syntax defines a syntax tree of several types. At the top-level there
is a StyleSheet. The style sheet is simply a collection of Rules. A Rule can be
either an AtRule or a QualifiedRule.

An AtRule is defined as a rule starting with an "@" symbol and an identifier,
then it's followed by zero or more component values and finally ends with either
a {-block or a semicolon. The block is parsed simply as a collection of tokens
and it is up to the user to define the exact grammar.

A QualifiedRule is defined as a rule starting with one or more component values
and ending with a {-block.

Inside the {-blocks are a list of declarations. Despite the name, a list of
declarations can be either an AtRule or a Declaration. A Declaration is an
identifier followed by a colon followed by one or more component values. The
declaration can also have it's Important flag set if the last two non-whitespace
tokens are a case-insensitive "!important".

ComponentValues are the basic unit inside rules and declarations. A
ComponentValue can be either a SimpleBlock, a Function, or a Token. A simple
block starts with either a {, [, or (, has zero or more component values, and
then ends with the mirror of the starting token (}, ], or )). A Function is
an identifier immediately followed by a left parenthesis, then zero or more
component values, and then ending with a right parenthesis.


*/
package css
