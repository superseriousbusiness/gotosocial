package html

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

type AST struct {
	Children []*Tag
	Text     []byte
}

func (ast *AST) String() string {
	sb := strings.Builder{}
	for i, child := range ast.Children {
		if i != 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(child.ASTString())
	}
	return sb.String()
}

type Attr struct {
	Key, Val []byte
}

func (attr *Attr) String() string {
	return fmt.Sprintf(`%s="%s"`, string(attr.Key), string(attr.Val))
}

type Tag struct {
	Root       *AST
	Parent     *Tag
	Prev, Next *Tag
	Children   []*Tag
	Index      int

	Name               []byte
	Attrs              []Attr
	textStart, textEnd int
}

func (tag *Tag) getAttr(key []byte) ([]byte, bool) {
	for _, attr := range tag.Attrs {
		if bytes.Equal(key, attr.Key) {
			return attr.Val, true
		}
	}
	return nil, false
}

func (tag *Tag) GetAttr(key string) (string, bool) {
	val, ok := tag.getAttr([]byte(key))
	return string(val), ok
}

func (tag *Tag) Text() string {
	return string(tag.Root.Text[tag.textStart:tag.textEnd])
}

func (tag *Tag) String() string {
	sb := strings.Builder{}
	sb.WriteString("<")
	sb.Write(tag.Name)
	for _, attr := range tag.Attrs {
		sb.WriteString(" ")
		sb.WriteString(attr.String())
	}
	sb.WriteString(">")
	return sb.String()
}

func (tag *Tag) ASTString() string {
	sb := strings.Builder{}
	sb.WriteString(tag.String())
	for _, child := range tag.Children {
		sb.WriteString("\n  ")
		s := child.ASTString()
		s = strings.ReplaceAll(s, "\n", "\n  ")
		sb.WriteString(s)
	}
	return sb.String()
}

func Parse(r *parse.Input) (*AST, error) {
	ast := &AST{}
	root := &Tag{}
	cur := root

	l := NewLexer(r)
	for {
		tt, data := l.Next()
		switch tt {
		case ErrorToken:
			if err := l.Err(); err != io.EOF {
				return nil, err
			}
			ast.Children = root.Children
			return ast, nil
		case TextToken:
			ast.Text = append(ast.Text, data...)
		case StartTagToken:
			child := &Tag{
				Root:      ast,
				Parent:    cur,
				Index:     len(cur.Children),
				Name:      l.Text(),
				textStart: len(ast.Text),
			}
			if 0 < len(cur.Children) {
				child.Prev = cur.Children[len(cur.Children)-1]
				child.Prev.Next = child
			}
			cur.Children = append(cur.Children, child)
			cur = child
		case AttributeToken:
			val := l.AttrVal()
			if 0 < len(val) && (val[0] == '"' || val[0] == '\'') {
				val = val[1 : len(val)-1]
			}
			cur.Attrs = append(cur.Attrs, Attr{l.AttrKey(), val})
		case StartTagCloseToken:
			if voidTags[string(cur.Name)] {
				cur.textEnd = len(ast.Text)
				cur = cur.Parent
			}
		case EndTagToken, StartTagVoidToken:
			start := cur
			for start != root && !bytes.Equal(l.Text(), start.Name) {
				start = start.Parent
			}
			if start == root {
				// ignore
			} else {
				parent := start.Parent
				for cur != parent {
					cur.textEnd = len(ast.Text)
					cur = cur.Parent
				}
			}
		}
	}
}

func (ast *AST) Query(s string) (*Tag, error) {
	sel, err := ParseSelector(s)
	if err != nil {
		return nil, err
	}

	for _, child := range ast.Children {
		if match := child.query(sel); match != nil {
			return match, nil
		}
	}
	return nil, nil
}

func (tag *Tag) query(sel selector) *Tag {
	if sel.AppliesTo(tag) {
		return tag
	}
	for _, child := range tag.Children {
		if match := child.query(sel); match != nil {
			return match
		}
	}
	return nil
}

func (ast *AST) QueryAll(s string) ([]*Tag, error) {
	sel, err := ParseSelector(s)
	if err != nil {
		return nil, err
	}

	matches := []*Tag{}
	for _, child := range ast.Children {
		child.queryAll(&matches, sel)
	}
	return matches, nil
}

func (tag *Tag) queryAll(matches *[]*Tag, sel selector) {
	if sel.AppliesTo(tag) {
		*matches = append(*matches, tag)
	}
	for _, child := range tag.Children {
		child.queryAll(matches, sel)
	}
}

type attrSelector struct {
	op   byte // empty, =, ~, |
	attr []byte
	val  []byte
}

func (sel attrSelector) AppliesTo(tag *Tag) bool {
	val, ok := tag.getAttr(sel.attr)
	if !ok {
		return false
	}

	switch sel.op {
	case 0:
		return true
	case '=':
		return bytes.Equal(val, sel.val)
	case '~':
		if 0 < len(sel.val) {
			vals := bytes.Split(val, []byte(" "))
			for _, val := range vals {
				if bytes.Equal(val, sel.val) {
					return true
				}
			}
		}
	case '|':
		return bytes.Equal(val, sel.val) || bytes.HasPrefix(val, append(sel.val, '-'))
	}
	return false
}

func (attr attrSelector) String() string {
	sb := strings.Builder{}
	sb.Write(attr.attr)
	if attr.op != 0 {
		sb.WriteByte(attr.op)
		if attr.op != '=' {
			sb.WriteByte('=')
		}
		sb.WriteByte('"')
		sb.Write(attr.val)
		sb.WriteByte('"')
	}
	return sb.String()
}

type selectorNode struct {
	typ   []byte // is * for universal
	attrs []attrSelector
	op    byte // space or >, last is NULL
}

func (sel selectorNode) AppliesTo(tag *Tag) bool {
	if 0 < len(sel.typ) && !bytes.Equal(sel.typ, []byte("*")) && !bytes.Equal(sel.typ, tag.Name) {
		return false
	}
	for _, attr := range sel.attrs {
		if !attr.AppliesTo(tag) {
			return false
		}
	}
	return true
}

func (sel selectorNode) String() string {
	sb := strings.Builder{}
	sb.Write(sel.typ)
	for _, attr := range sel.attrs {
		if bytes.Equal(attr.attr, []byte("id")) && attr.op == '=' {
			sb.WriteByte('#')
			sb.Write(attr.val)
		} else if bytes.Equal(attr.attr, []byte("class")) && attr.op == '~' {
			sb.WriteByte('.')
			sb.Write(attr.val)
		} else {
			sb.WriteByte('[')
			sb.WriteString(attr.String())
			sb.WriteByte(']')
		}
	}
	if sel.op != 0 {
		sb.WriteByte(' ')
		sb.WriteByte(sel.op)
		sb.WriteByte(' ')
	}
	return sb.String()
}

type token struct {
	tt   css.TokenType
	data []byte
}

type selector []selectorNode

func ParseSelector(s string) (selector, error) {
	ts := []token{}
	l := css.NewLexer(parse.NewInputString(s))
	for {
		tt, data := l.Next()
		if tt == css.ErrorToken {
			if err := l.Err(); err != io.EOF {
				return selector{}, err
			}
			break
		}
		ts = append(ts, token{
			tt:   tt,
			data: data,
		})
	}

	sel := selector{}
	node := selectorNode{}
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		if 0 < i && (t.tt == css.WhitespaceToken || t.tt == css.DelimToken && t.data[0] == '>') {
			if t.tt == css.DelimToken {
				node.op = '>'
			} else {
				node.op = ' '
			}
			sel = append(sel, node)
			node = selectorNode{}
		} else if t.tt == css.IdentToken || t.tt == css.DelimToken && t.data[0] == '*' {
			node.typ = t.data
		} else if t.tt == css.DelimToken && (t.data[0] == '.' || t.data[0] == '#') && i+1 < len(ts) && ts[i+1].tt == css.IdentToken {
			if t.data[0] == '#' {
				node.attrs = append(node.attrs, attrSelector{op: '=', attr: []byte("id"), val: ts[i+1].data})
			} else {
				node.attrs = append(node.attrs, attrSelector{op: '~', attr: []byte("class"), val: ts[i+1].data})
			}
			i++
		} else if t.tt == css.DelimToken && t.data[0] == '[' && i+2 < len(ts) && ts[i+1].tt == css.IdentToken && ts[i+2].tt == css.DelimToken {
			if ts[i+2].data[0] == ']' {
				node.attrs = append(node.attrs, attrSelector{op: 0, attr: ts[i+1].data})
				i += 2
			} else if i+4 < len(ts) && ts[i+3].tt == css.IdentToken && ts[i+4].tt == css.DelimToken && ts[i+4].data[0] == ']' {
				node.attrs = append(node.attrs, attrSelector{op: ts[i+2].data[0], attr: ts[i+1].data, val: ts[i+3].data})
				i += 4
			}
		}
	}
	sel = append(sel, node)
	return sel, nil
}

func (sels selector) AppliesTo(tag *Tag) bool {
	if len(sels) == 0 {
		return true
	} else if !sels[len(sels)-1].AppliesTo(tag) {
		return false
	}

	tag = tag.Parent
	isel := len(sels) - 2
	for 0 <= isel && tag != nil {
		switch sels[isel].op {
		case ' ':
			for tag != nil {
				if sels[isel].AppliesTo(tag) {
					break
				}
				tag = tag.Parent
			}
		case '>':
			if !sels[isel].AppliesTo(tag) {
				return false
			}
			tag = tag.Parent
		default:
			return false
		}
		isel--
	}
	return len(sels) != 0 && isel == -1
}

func (sels selector) String() string {
	if len(sels) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for _, sel := range sels {
		sb.WriteString(sel.String())
	}
	return sb.String()[1:]
}

var voidTags = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"source": true,
	"track":  true,
	"wbr":    true,
}
