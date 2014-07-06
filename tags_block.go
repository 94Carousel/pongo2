package pongo2

import (
	"errors"
	"fmt"
)

type tagBlockNode struct {
	name string
}

func (node *tagBlockNode) getBlockWrapperByName(tpl *Template) *NodeWrapper {
	var t *NodeWrapper
	if tpl.child != nil {
		// First ask the child for the block
		t = node.getBlockWrapperByName(tpl.child)
	}
	if t == nil {
		// Child has no block, lets look up here at parent
		t = tpl.blocks[node.name]
	}
	return t
}

func (node *tagBlockNode) Execute(ctx *ExecutionContext) (string, error) {
	tpl := ctx.template
	if tpl == nil {
		panic("internal error: tpl == nil")
	}
	// Determine the block to execute
	block_wrapper := node.getBlockWrapperByName(tpl)
	if block_wrapper == nil {
		fmt.Printf("could not find: %s\n", node.name)
		return "", errors.New("boo")
	}
	rv, err := block_wrapper.Execute(ctx)
	if err != nil {
		return "", err
	}

	// TODO: Add support for {{ block.super }}

	return rv, nil
}

func tagBlockParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'block' requires an identifier.", nil)
	}

	name_token := arguments.MatchType(TokenIdentifier)
	if name_token == nil {
		return nil, arguments.Error("First argument for tag 'block' must be an identifier.", nil)
	}

	if arguments.Remaining() != 0 {
		return nil, arguments.Error("Tag 'block' takes exactly 1 argument (an identifier).", nil)
	}

	wrapper, endtagargs, err := doc.WrapUntilTag("endblock")
	if err != nil {
		return nil, err
	}
	for endtagargs.Remaining() > 0 {
		if endtagname_token := endtagargs.MatchType(TokenIdentifier); endtagname_token != nil {
			if endtagname_token.Val != name_token.Val {
				return nil, endtagargs.Error(fmt.Sprintf("Name for 'endblock' must equal to 'block'-tag's name ('%s' != '%s').",
					name_token.Val, endtagname_token.Val), nil)
			}
		} else {
			return nil, endtagargs.Error("Either no or only one argument (identifier) allowed for 'endblock'.", nil)
		}
	}

	tpl := doc.template
	if tpl == nil {
		panic("internal error: tpl == nil")
	}
	_, has_block := tpl.blocks[name_token.Val]
	if !has_block {
		tpl.blocks[name_token.Val] = wrapper
	} else {
		return nil, arguments.Error(fmt.Sprintf("Block named '%s' already defined", name_token.Val), nil)
	}

	return &tagBlockNode{name: name_token.Val}, nil
}

func init() {
	RegisterTag("block", tagBlockParser)
}
