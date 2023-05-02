// Package ast defines the abstract syntax tree of mybatis mapper xml.
package ast

import (
	"encoding/xml"
	"io"
	"strings"
)

var (
	_ Node = (*IfNode)(nil)

	_ Node = (*ChooseNode)(nil)
	_ Node = (*WhenNode)(nil)
	_ Node = (*OtherwiseNode)(nil)

	_ Node = (*WhereNode)(nil)
	_ Node = (*SetNode)(nil)
	_ Node = (*TrimNode)(nil)
)

// IfNode represents a if node in mybatis mapper xml likes <if test="condition">...</if>.
type IfNode struct {
	Test     string
	Children []Node
}

// NewIfNode creates a new if node.
func NewIfNode(startElement *xml.StartElement) *IfNode {
	node := &IfNode{}
	for _, attr := range startElement.Attr {
		if attr.Name.Local == "test" {
			node.Test = attr.Value
		}
	}
	return node
}

// RestoreSQL implements Node interface, the if condition will be ignored.
func (n *IfNode) RestoreSQL(w io.Writer) error {
	if len(n.Children) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
	}
	for _, node := range n.Children {
		if err := node.RestoreSQL(w); err != nil {
			return err
		}
	}
	return nil
}

// AddChild adds a child to the if node.
func (n *IfNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// ChooseNode represents a choose node in mybatis mapper xml likes <choose>...</choose>.
type ChooseNode struct {
	Children []Node
}

// NewChooseNode creates a new choose node.
func NewChooseNode(_ *xml.StartElement) *ChooseNode {
	return &ChooseNode{}
}

// RestoreSQL implements Node interface.
func (n *ChooseNode) RestoreSQL(w io.Writer) error {
	if len(n.Children) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
	}
	for _, node := range n.Children {
		if err := node.RestoreSQL(w); err != nil {
			return err
		}
	}
	return nil
}

// AddChild implements Node interface.
func (n *ChooseNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// WhenNode represents a when node in mybatis mapper xml select node likes <select><when test="condition">...</when></select>.
type WhenNode struct {
	Test     string
	Children []Node
}

// NewWhenNode creates a new when node.
func NewWhenNode(startElement *xml.StartElement) *WhenNode {
	node := &WhenNode{}
	for _, attr := range startElement.Attr {
		if attr.Name.Local == "test" {
			node.Test = attr.Value
		}
	}
	return node
}

// RestoreSQL implements Node interface, the when condition will be ignored.
func (n *WhenNode) RestoreSQL(w io.Writer) error {
	if len(n.Children) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
	}
	if len(n.Children) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
	}
	for _, node := range n.Children {
		if err := node.RestoreSQL(w); err != nil {
			return err
		}
	}
	return nil
}

// AddChild adds a child to the when node.
func (n *WhenNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// OtherwiseNode represents a otherwise node in mybatis mapper xml select node likes <select><otherwise>...</otherwise></select>.
type OtherwiseNode struct {
	Children []Node
}

// NewOtherwiseNode creates a new otherwise node.
func NewOtherwiseNode(_ *xml.StartElement) *OtherwiseNode {
	return &OtherwiseNode{}
}

// RestoreSQL implements Node interface.
func (n *OtherwiseNode) RestoreSQL(w io.Writer) error {
	if len(n.Children) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
	}
	for _, node := range n.Children {
		if err := node.RestoreSQL(w); err != nil {
			return err
		}
	}
	return nil
}

// AddChild adds a child to the otherwise node.
func (n *OtherwiseNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// TrimNode represents a trim node in mybatis mapper xml likes <trim prefix="prefix" suffix="suffix" prefixOverrides="prefixOverrides" suffixOverrides="suffixOverrides">...</trim>.
type TrimNode struct {
	Prefix               string
	Suffix               string
	PrefixOverridesParts []string
	SuffixOverridesParts []string
	Children             []Node
}

// NewTrimNode creates a new trim node.
func NewTrimNode(startElement *xml.StartElement) *TrimNode {
	var prefix, suffix, prefixOverrides, suffixOverrides string
	for _, attr := range startElement.Attr {
		switch attr.Name.Local {
		case "prefix":
			prefix = attr.Value
		case "suffix":
			suffix = attr.Value
		case "prefixOverrides":
			prefixOverrides = attr.Value
		case "suffixOverrides":
			suffixOverrides = attr.Value
		}
	}
	return newTrimNodeWithAttrs(prefix, suffix, prefixOverrides, suffixOverrides)
}

// newTrimNodeWithAttrs creates a new trim node with given attributes.
func newTrimNodeWithAttrs(prefix, suffix, prefixOverrides, suffixOverrides string) *TrimNode {
	prefixOverridesParts := strings.Split(prefixOverrides, "|")
	suffixOverridesParts := strings.Split(suffixOverrides, "|")
	return &TrimNode{
		Prefix:               prefix,
		Suffix:               suffix,
		PrefixOverridesParts: prefixOverridesParts,
		SuffixOverridesParts: suffixOverridesParts,
	}
}

// RestoreSQL implements Node interface.
func (n *TrimNode) RestoreSQL(w io.Writer) error {
	var stringsBuilder strings.Builder
	for _, node := range n.Children {
		if err := node.RestoreSQL(&stringsBuilder); err != nil {
			return err
		}
	}
	trimmed := strings.TrimSpace(stringsBuilder.String())
	if len(trimmed) == 0 {
		return nil
	}
	// Replace the prefix and suffix with empty string if matches the part in prefixOverridesParts and suffixOverridesParts.
	for _, part := range n.PrefixOverridesParts {
		if strings.HasPrefix(trimmed, part) {
			trimmed = strings.TrimPrefix(trimmed, part)
			break
		}
	}
	for _, part := range n.SuffixOverridesParts {
		if strings.HasSuffix(trimmed, part) {
			trimmed = strings.TrimSuffix(trimmed, part)
			break
		}
	}
	if len(n.Prefix) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(n.Prefix)); err != nil {
			return err
		}
	}
	if len(trimmed) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(trimmed)); err != nil {
			return err
		}
	}
	if len(n.Suffix) > 0 {
		if _, err := w.Write([]byte(" ")); err != nil {
			return err
		}
		if _, err := w.Write([]byte(n.Suffix)); err != nil {
			return err
		}
	}
	return nil
}

// AddChild adds a child to the trim node.
func (n *TrimNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// WhereNode represents a where node in mybatis mapper xml likes <where>...</where>.
type WhereNode struct {
	trimNode *TrimNode
}

// NewWhereNode creates a new where node.
func NewWhereNode(_ *xml.StartElement) *WhereNode {
	return &WhereNode{
		trimNode: newTrimNodeWithAttrs("WHERE", "", "AND |OR ", ""),
	}
}

// RestoreSQL implements Node interface.
func (n *WhereNode) RestoreSQL(w io.Writer) error {
	return n.trimNode.RestoreSQL(w)
}

// AddChild adds a child to the where node.
func (n *WhereNode) AddChild(child Node) {
	n.trimNode.AddChild(child)
}

// SetNode represents a set node in mybatis mapper xml likes <set>...</set>.
type SetNode struct {
	trimNode *TrimNode
}

// NewSetNode creates a new set node.
func NewSetNode(_ *xml.StartElement) *SetNode {
	return &SetNode{
		trimNode: newTrimNodeWithAttrs("SET", "", "", ","),
	}
}

// RestoreSQL implements Node interface.
func (n *SetNode) RestoreSQL(w io.Writer) error {
	return n.trimNode.RestoreSQL(w)
}

// AddChild adds a child to the set node.
func (n *SetNode) AddChild(child Node) {
	n.trimNode.AddChild(child)
}
