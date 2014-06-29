package pongo2

// Doc = { ( Filter | Tag | HTML ) }
func (p *Parser) parseDocElement() (INode, error) {
	t := p.Current()

	switch t.Typ {
	case TokenHTML:
		p.Consume() // consume HTML element
		return &NodeHTML{token: t}, nil
	case TokenSymbol:
		switch t.Val {
		case "{{":
			// parse variable
			variable, err := p.parseVariableElement()
			if err != nil {
				return nil, err
			}
			return variable, nil
		case "{%":
			// parse tag
			tag, err := p.parseTagElement()
			if err != nil {
				return nil, err
			}
			return tag, nil
		}
	}
	return nil, p.Error("Unexpected token (only HTML/tags/filters in templates allowed)", t)
}

func (tpl *Template) parse() error {
	tpl.parser = newParser(tpl.name, tpl.tokens, tpl)
	doc, err := tpl.parser.parseDocument()
	if err != nil {
		return err
	}
	tpl.root = doc
	return nil
}

func (p *Parser) parseDocument() (*NodeDocument, error) {
	doc := &NodeDocument{}

	for p.Remaining() > 0 {
		node, err := p.parseDocElement()
		if err != nil {
			return nil, err
		}
		doc.Nodes = append(doc.Nodes, node)
	}

	return doc, nil
}
