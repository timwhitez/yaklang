package formatter

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"strings"
)

type RuleNode struct {
	RuleId      int
	Parent      *RuleNode
	Data        []any // antlr.Token、RuleNode
	DataVerbose []string
}

func NewRuleNode(id int) *RuleNode {
	return &RuleNode{
		RuleId:      id,
		Parent:      nil,
		Data:        []any{},
		DataVerbose: []string{},
	}
}

func (t *RuleNode) AppendNode(d any) {
	t.Data = append(t.Data, d)
	if v, ok := d.(*RuleNode); ok {
		v.Parent = t
		t.DataVerbose = append(t.DataVerbose, fmt.Sprint(v.RuleId))
	}
	if v, ok := d.(antlr.Token); ok {
		t.DataVerbose = append(t.DataVerbose, v.GetText())
	}
}

type ParseTreeListener struct {
	tokenTreeRoot *RuleNode
	currentNode   *RuleNode
	preTokenIndex int
	tokenStream   *antlr.CommonTokenStream
	antlr.BaseParseTreeListener
}

func (p *ParseTreeListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	for ; p.preTokenIndex < ctx.GetStart().GetTokenIndex(); p.preTokenIndex++ {
		p.currentNode.AppendNode(p.tokenStream.Get(p.preTokenIndex))
	}
	newNode := NewRuleNode(ctx.GetRuleIndex())
	p.currentNode.AppendNode(newNode)
	p.currentNode = newNode
}
func (p *ParseTreeListener) ExitEveryRule(ctx antlr.ParserRuleContext) {
	for ; p.preTokenIndex < ctx.GetStop().GetTokenIndex()+1; p.preTokenIndex++ {
		p.currentNode.AppendNode(p.tokenStream.Get(p.preTokenIndex))
	}
	p.currentNode = p.currentNode.Parent
}

type joinWith string
type template string

const (
	space   joinWith = " "
	newline joinWith = "\n"
)

func Format(code string) (formattedCode string) {
	formattedCode = code
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("format code failed: %s", err)
		}
	}()
	lexer := yak.NewYaklangLexer(antlr.NewInputStream(code))
	lexer.RemoveErrorListeners()
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	tokenStream.Fill()
	parser := yak.NewYaklangParser(tokenStream)
	parser.RemoveErrorListeners()
	raw := parser.Program()
	root := NewRuleNode(yak.YaklangParserRULE_program)
	listener := &ParseTreeListener{tokenStream: tokenStream, tokenTreeRoot: root, currentNode: root, preTokenIndex: 0}
	antlr.ParseTreeWalkerDefault.Walk(listener, raw)
	blockN := 0
	var walkTree func(node *RuleNode) string
	walkTree = func(node *RuleNode) string {
		var codeItems []string
		for _, v := range node.Data {
			if token, ok := v.(antlr.Token); ok {
				codeItems = append(codeItems, token.GetText())
			} else if code, ok := v.(string); ok {
				codeItems = append(codeItems, code)
			} else if ruleNode, ok := v.(*RuleNode); ok {
				if ruleNode.RuleId == yak.YaklangParserRULE_block {
					blockN++
				}
				codeItems = append(codeItems, walkTree(ruleNode))
				if ruleNode.RuleId == yak.YaklangParserRULE_block {
					blockN--
				}
			}
		}
		res := ""
		// 处理了注释和逗号运算符
		formatterWithSpace := func(sep string) {
			for i, item := range codeItems {
				if token, ok := node.Data[i].(antlr.Token); ok {
					if token.GetTokenType() == yak.YaklangParserComma {
						res += item
						continue
					}
					if token.GetTokenType() == yak.YaklangParserCOMMENT {
						if len(res) > 0 && res[len(res)-1:] != " " {
							res += " "
						}
						res += item
						continue
					}
					if token.GetTokenType() == yak.YaklangParserLINE_COMMENT {
						if len(res) > 0 && res[len(res)-1:] != " " {
							res += " "
						}
						res += item
						continue
					}

				}
				if len(res) > 0 && res[len(res)-1:] != sep {
					res += sep
				}
				res += item

			}
		}

		switch node.RuleId {
		case yak.YaklangParserRULE_statementList:
			formatterWithSpace("\n")
			res = strings.TrimSpace(res)
			// assignExpression、expressionList
		case yak.YaklangParserRULE_declareAndAssignExpression, yak.YaklangParserRULE_declareVariableOnly, yak.YaklangParserRULE_leftExpressionList, yak.YaklangParserRULE_expressionList, yak.YaklangParserRULE_tryStmt, yak.YaklangParserRULE_ifStmt:
			formatterWithSpace(" ")
		case yak.YaklangParserRULE_functionCall:
			formatterWithSpace("")
		case yak.YaklangParserRULE_ordinaryArguments:
			formatterWithSpace(" ")
		case yak.YaklangParserRULE_expression:
			spaceFlag := []string{"!", "-", "+", "^", "&", "*", "("}
			hasFlag := false
			for _, flag := range spaceFlag {
				if utils.StringArrayContains(codeItems, flag) {
					hasFlag = true
				}
			}
			idFlag := []int{yak.YaklangParserRULE_memberCall, yak.YaklangParserRULE_sliceCall, yak.YaklangParserRULE_functionCall}
			for _, datum := range node.Data {
				for _, flag := range idFlag {
					if v, ok := datum.(*RuleNode); ok && v.RuleId == flag {
						hasFlag = true
					}
				}
			}
			if hasFlag {
				formatterWithSpace("")
			} else {
				formatterWithSpace(" ")
			}
		//case yak.YaklangParserRULE_statement:
		//	formatterWithSpace(" ")
		case yak.YaklangParserRULE_assignExpression:
			spaceflag := []string{"++", "--"}
			hasFlag := false
			for _, flag := range spaceflag {
				if utils.StringArrayContains(codeItems, flag) {
					hasFlag = true
				}
			}
			if hasFlag {
				formatterWithSpace("")
			} else {
				formatterWithSpace(" ")
			}
		case yak.YaklangParserRULE_block:
			codes := strings.Join(codeItems[1:len(codeItems)-1], "")
			codes = strings.TrimSpace(codes) // 由于每个statement末尾都有\n，在split前需要去掉最后一个
			splits := strings.Split(codes, "\n")
			res += codeItems[0] + "\n"
			res += "\t" + strings.Join(splits, "\n\t") + "\n"
			res += codeItems[len(codeItems)-1]
		default:
			res += strings.Join(codeItems, "")
		}
		return res
	}
	return walkTree(root)
}
