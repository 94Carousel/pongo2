package pongo2

import (
	"fmt"
	"math"
)

type Expression struct {
	expr1    IEvaluator
	expr2    IEvaluator
	op_token *Token
}

type relationalExpression struct {
	expr1    IEvaluator
	expr2    IEvaluator
	op_token *Token
}

type simpleExpression struct {
	negate        bool
	negative_sign bool
	term1         IEvaluator
	term2         IEvaluator
	op_token      *Token
}

type term struct {
	factor1  IEvaluator
	factor2  IEvaluator
	op_token *Token
}

type power struct {
	power1 IEvaluator
	power2 IEvaluator
}

func (expr *Expression) Execute(ctx *ExecutionContext) (string, error) {
	value, err := expr.Evaluate(ctx)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

func (expr *Expression) Evaluate(ctx *ExecutionContext) (*Value, error) {
	v1, err := expr.expr1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if expr.expr2 != nil {
		v2, err := expr.expr2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.op_token.Val {
		case "and", "&&":
			return AsValue(v1.IsTrue() && v2.IsTrue()), nil
		case "or", "||":
			return AsValue(v1.IsTrue() || v2.IsTrue()), nil
		default:
			panic(fmt.Sprintf("unimplemented: %s", expr.op_token.Val))
		}
	} else {
		return v1, nil
	}
}

func (expr *relationalExpression) Evaluate(ctx *ExecutionContext) (*Value, error) {
	v1, err := expr.expr1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if expr.expr2 != nil {
		v2, err := expr.expr2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.op_token.Val {
		case "<=":
			return AsValue(v1.Integer() <= v2.Integer()), nil
		case ">=":
			return AsValue(v1.Integer() >= v2.Integer()), nil
		case "==":
			return AsValue(v1.EqualValueTo(v2)), nil
		case ">":
			return AsValue(v1.Integer() > v2.Integer()), nil
		case "<":
			return AsValue(v1.Integer() < v2.Integer()), nil
		case "!=", "<>":
			return AsValue(!v1.EqualValueTo(v2)), nil
		case "in":
			return AsValue(v2.Contains(v1)), nil
		default:
			panic(fmt.Sprintf("unimplemented: %s", expr.op_token.Val))
		}
	} else {
		return v1, nil
	}
}

func (expr *simpleExpression) Evaluate(ctx *ExecutionContext) (*Value, error) {
	t1, err := expr.term1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	result := t1

	if expr.negate || expr.negative_sign {
		result = result.Negate()
	}

	if expr.term2 != nil {
		t2, err := expr.term2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch expr.op_token.Val {
		case "+":
			return AsValue(result.Integer() + t2.Integer()), nil
		case "-":
			return AsValue(result.Integer() - t2.Integer()), nil
		default:
			panic("unimplemented")
		}
	}

	return result, nil
}

func (t *term) Evaluate(ctx *ExecutionContext) (*Value, error) {
	f1, err := t.factor1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if t.factor2 != nil {
		f2, err := t.factor2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		switch t.op_token.Val {
		case "*":
			return AsValue(f1.Integer() * f2.Integer()), nil
		case "/":
			panic("unimplemented")
		default:
			panic("unimplemented")
		}
	} else {
		return f1, nil
	}
}

func (pw *power) Evaluate(ctx *ExecutionContext) (*Value, error) {
	p1, err := pw.power1.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	if pw.power2 != nil {
		p2, err := pw.power2.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		return AsValue(math.Pow(p1.Float(), p2.Float())), nil
	} else {
		return p1, nil
	}
}

func (p *Parser) parseFactor() (IEvaluator, error) {
	if p.Match(TokenSymbol, "(") != nil {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if p.Match(TokenSymbol, ")") == nil {
			return nil, p.Error("Closing bracket expected after expression", nil)
		}
		return expr, nil
	}

	return p.parseVariableOrLiteralWithFilter()
}

func (p *Parser) parseTerm() (IEvaluator, error) {
	term := new(term)

	factor1, err := p.parsePower()
	if err != nil {
		return nil, err
	}
	term.factor1 = factor1

	if p.PeekOne(TokenSymbol, "*", "/") != nil {
		op := p.Current()
		p.Consume()
		factor2, err := p.parsePower()
		if err != nil {
			return nil, err
		}
		term.factor2 = factor2
		term.op_token = op
	}

	return term, nil
}

func (p *Parser) parsePower() (IEvaluator, error) {
	pw := new(power)

	power1, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	pw.power1 = power1

	if p.Match(TokenSymbol, "^") != nil {
		power2, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		pw.power2 = power2
	}

	return pw, nil
}

func (p *Parser) parseSimpleExpression() (IEvaluator, error) {
	expr := new(simpleExpression)

	if sign := p.MatchOne(TokenSymbol, "+", "-"); sign != nil {
		if sign.Val == "-" {
			expr.negative_sign = true
		}
	}

	if p.Match(TokenSymbol, "!") != nil {
		expr.negate = true
	}

	term1, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	expr.term1 = term1

	if p.PeekOne(TokenSymbol, "+", "-") != nil {
		op := p.Current()
		p.Consume()
		term2, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		expr.term2 = term2
		expr.op_token = op
	}

	return expr, nil
}

func (p *Parser) parseRelationalExpression() (IEvaluator, error) {
	expr1, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}

	expr := &relationalExpression{
		expr1: expr1,
	}

	if t := p.MatchOne(TokenSymbol, "==", "<=", ">=", "!=", "<>", ">", "<"); t != nil {
		expr2, err := p.parseRelationalExpression()
		if err != nil {
			return nil, err
		}
		expr.op_token = t
		expr.expr2 = expr2
	} else if t := p.MatchOne(TokenKeyword, "in"); t != nil {
		expr2, err := p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
		expr.op_token = t
		expr.expr2 = expr2
	}

	return expr, nil
}

func (p *Parser) ParseExpression() (INodeEvaluator, error) {
	rexpr1, err := p.parseRelationalExpression()
	if err != nil {
		return nil, err
	}

	exp := &Expression{
		expr1: rexpr1,
	}

	if p.PeekOne(TokenSymbol, "&&", "||") != nil || p.PeekOne(TokenKeyword, "and", "or") != nil {
		op := p.Current()
		p.Consume()
		expr2, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		exp.expr2 = expr2
		exp.op_token = op
	}

	return exp, nil
}
