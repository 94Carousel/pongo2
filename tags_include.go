package pongo2

import (
	"errors"
	"path/filepath"
)

type tagIncludeNode struct {
	tpl                *Template
	filename_evaluator IEvaluator
	lazy               bool
	only               bool
	filename           string
	with_pairs         map[string]IEvaluator
}

func (node *tagIncludeNode) Execute(ctx *ExecutionContext) (string, error) {
	// Building the context for the template
	include_ctx := make(Context)

	// Fill the context with all data from the parent
	if !node.only {
		for key, value := range ctx.Public {
			include_ctx[key] = value
		}
	}

	// Put all custom with-pairs into the context
	for key, value := range node.with_pairs {
		val, err := value.Evaluate(ctx)
		if err != nil {
			return "", err
		}
		include_ctx[key] = val
	}

	// Execute the template
	if node.lazy {
		// Evaluate the filename
		filename, err := node.filename_evaluator.Evaluate(ctx)
		if err != nil {
			return "", err
		}

		if filename.String() == "" {
			return "", errors.New("Filename for 'include'-tag evaluated to an empty string.")
		}

		// Get include-filename relative to the including-template directory
		including_dir := filepath.Dir(ctx.template.name)
		included_filename := filepath.Join(including_dir, filename.String())

		included_tpl, err := FromFile(included_filename)
		if err != nil {
			return "", err
		}
		return included_tpl.Execute(include_ctx)
	} else {
		// Template is already parsed with static filename
		return node.tpl.Execute(include_ctx)
	}
}

func tagIncludeParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	include_node := &tagIncludeNode{
		with_pairs: make(map[string]IEvaluator),
	}

	if filename_token := arguments.MatchType(TokenString); filename_token != nil {
		// prepared, static template

		// Get include-filename relative to the including-template directory
		including_dir := filepath.Dir(doc.template.name)
		included_filename := filepath.Join(including_dir, filename_token.Val)

		// Parse the parent
		include_node.filename = included_filename
		included_tpl, err := FromFile(included_filename)
		if err != nil {
			return nil, err
		}
		include_node.tpl = included_tpl
	} else {
		// No String, then the user wants to use lazy-evaluation (slower, but possible)
		filename_evaluator, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		include_node.filename_evaluator = filename_evaluator
		include_node.lazy = true
	}

	// After having parsed the filename we're gonna parse the with+only options
	if arguments.Match(TokenIdentifier, "with") != nil {
		for arguments.Remaining() > 0 {
			// We have at least one key=expr pair (because of starting "with")
			key_token := arguments.MatchType(TokenIdentifier)
			if key_token == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			value_expr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}

			include_node.with_pairs[key_token.Val] = value_expr

			// Only?
			if arguments.Match(TokenIdentifier, "only") != nil {
				include_node.only = true
				break // stop parsing arguments because it's the last option
			}
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed 'include'-tag arguments.", nil)
	}

	return include_node, nil
}

func init() {
	RegisterTag("include", tagIncludeParser)
}
