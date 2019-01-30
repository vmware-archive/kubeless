/*
Copyright 2018 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package ast provides AST nodes and ancillary structures and algorithms.
package ast

import (
	"fmt"
)

// Fodder

// FodderKind is an enum.
type FodderKind int

const (
	// FodderLineEnd represents a line ending.
	//
	// It indicates that the next token, paragraph, or interstitial
	// should be on a new line.
	//
	// A single comment string is allowed, which flows before the new line.
	//
	// The LineEnd fodder specifies the indentation level and vertical spacing
	// before whatever comes next.
	FodderLineEnd FodderKind = iota

	// FodderInterstitial represents a comment in middle of a line.
	//
	// They must be /* C-style */ comments.
	//
	// If it follows a token (i.e., it is the first fodder element) then it
	// appears after the token on the same line.  If it follows another
	// interstitial, it will also flow after it on the same line.  If it follows
	// a new line or a paragraph, it is the first thing on the following line,
	// after the blank lines and indentation specified by the previous fodder.
	//
	// There is exactly one comment string.
	FodderInterstitial

	// FodderParagraph represents a comment consisting of at least one line.
	//
	// // and # style comments have exactly one line.  C-style comments can have
	// more than one line.
	//
	// All lines of the comment are indented according to the indentation level
	// of the previous new line / paragraph fodder.
	//
	// The Paragraph fodder specifies the indentation level and vertical spacing
	// before whatever comes next.
	FodderParagraph
)

// FodderElement is a single piece of fodder.
type FodderElement struct {
	Kind    FodderKind
	Blanks  int
	Indent  int
	Comment []string
}

// MakeFodderElement is a helper function that checks some preconditions.
func MakeFodderElement(kind FodderKind, blanks int, indent int, comment []string) FodderElement {
	if kind == FodderLineEnd && len(comment) > 1 {
		panic(fmt.Sprintf("FodderLineEnd but comment == %v.", comment))
	}
	if kind == FodderInterstitial && blanks > 0 {
		panic(fmt.Sprintf("FodderInterstitial but blanks == %d", blanks))
	}
	if kind == FodderInterstitial && indent > 0 {
		panic(fmt.Sprintf("FodderInterstitial but indent == %d", blanks))
	}
	if kind == FodderInterstitial && len(comment) != 1 {
		panic(fmt.Sprintf("FodderInterstitial but comment == %v.", comment))
	}
	if kind == FodderParagraph && len(comment) == 0 {
		panic(fmt.Sprintf("FodderParagraph but comment was empty"))
	}
	return FodderElement{Kind: kind, Blanks: blanks, Indent: indent, Comment: comment}
}

// Fodder is stuff that is usually thrown away by lexers/preprocessors but is
// kept so that the source can be round tripped with near full fidelity.
type Fodder []FodderElement

// FodderHasCleanEndline is true if the fodder doesn't end with an interstitial.
func FodderHasCleanEndline(fodder Fodder) bool {
	return len(fodder) > 0 && fodder[len(fodder)-1].Kind != FodderInterstitial
}

// FodderAppend appends to the fodder but preserves constraints.
//
// See FodderConcat below.
func FodderAppend(a *Fodder, elem FodderElement) {
	if FodderHasCleanEndline(*a) && elem.Kind == FodderLineEnd {
		if len(elem.Comment) > 0 {
			// The line end had a comment, so create a single line paragraph for it.
			*a = append(*a, MakeFodderElement(FodderParagraph, elem.Blanks, elem.Indent, elem.Comment))
		} else {
			back := &(*a)[len(*a)-1]
			// Merge it into the previous line end.
			back.Indent = elem.Indent
			back.Blanks += elem.Blanks
		}
	} else {
		if !FodderHasCleanEndline(*a) && elem.Kind == FodderParagraph {
			*a = append(*a, MakeFodderElement(FodderLineEnd, 0, elem.Indent, []string{}))
		}
		*a = append(*a, elem)
	}
}

// FodderConcat concats the two fodders but also preserves constraints.
//
// Namely, a FodderLineEnd is not allowed to follow a FodderParagraph or a FodderLineEnd.
func FodderConcat(a Fodder, b Fodder) Fodder {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	r := a
	// Carefully add the first element of b.
	FodderAppend(&r, b[0])
	// Add the rest of b.
	for i := 1; i < len(b); i++ {
		r = append(r, b[i])
	}
	return r
}

// FodderMoveFront moves b to the front of a.
func FodderMoveFront(a *Fodder, b *Fodder) {
	*a = FodderConcat(*b, *a)
	*b = Fodder{}
}

// FodderEnsureCleanNewline adds a LineEnd to the fodder if necessary.
func FodderEnsureCleanNewline(fodder *Fodder) {
	if !FodderHasCleanEndline(*fodder) {
		FodderAppend(fodder, MakeFodderElement(FodderLineEnd, 0, 0, []string{}))
	}
}

// FodderElementCountNewlines returns the number of new line chars represented by the fodder element
func FodderElementCountNewlines(elem FodderElement) int {
	switch elem.Kind {
	case FodderInterstitial:
		return 0
	case FodderLineEnd:
		return 1
	case FodderParagraph:
		return len(elem.Comment) + elem.Blanks
	}
	panic(fmt.Sprintf("Unknown FodderElement kind %d", elem.Kind))
}

// FodderCountNewlines returns the number of new line chars represented by the fodder.
func FodderCountNewlines(fodder Fodder) int {
	sum := 0
	for _, elem := range fodder {
		sum += FodderElementCountNewlines(elem)
	}
	return sum
}
