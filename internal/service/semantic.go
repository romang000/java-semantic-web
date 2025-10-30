package service

import (
	"encoding/json"
	"fmt"
	
	"github.com/romang000/java-semantic-web/internal/models"
)

type SemanticError struct {
	Message  string
	Position map[string]int
}

func (s SemanticError) Error() string {
	if s.Position != nil {
		return fmt.Sprintf("%s at %v", s.Message, s.Position)
	}
	return s.Message
}

type Symbol struct {
	Name string
	Type string
}

type Scope struct {
	symbols map[string]*Symbol
	parent  *Scope
}

type Analyzer struct {
	root          *models.AstNode
	scope         *Scope
	errors        []SemanticError
	currentMethod *models.AstNode
}

func NewScope(parent *Scope) *Scope {
	return &Scope{symbols: map[string]*Symbol{}, parent: parent}
}

func (s *Scope) Define(sym *Symbol) error {
	if _, exists := s.symbols[sym.Name]; exists {
		return fmt.Errorf("symbol '%s' already defined", sym.Name)
	}
	s.symbols[sym.Name] = sym
	return nil
}

func (s *Scope) Resolve(name string) (*Symbol, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if sym, ok := sc.symbols[name]; ok {
			return sym, true
		}
	}
	return nil, false
}

func NewAnalyzer(root *models.AstNode) *Analyzer {
	return &Analyzer{
		root:  root,
		scope: NewScope(nil),
	}
}

func (a *Analyzer) Analyze() ([]SemanticError, *models.AstNode) {
	a.visitNode(a.root)
	if len(a.errors) > 0 {
		return a.errors, nil
	}
	return nil, a.root
}

func (a *Analyzer) addError(message string, position map[string]int) {
	a.errors = append(a.errors, SemanticError{Message: message, Position: position})
}

func (a *Analyzer) visitNode(n *models.AstNode) {
	if n == nil {
		return
	}
	
	switch n.NodeType {
	
	case "Program":
		for _, cls := range n.Classes {
			a.visitNode(cls)
		}
	
	case "Package", "Import":
		// poxyi
	
	case "ClassDeclaration":
		a.visitClass(n)
	
	case "FieldDeclaration":
		a.visitVarDecl(n)
	
	case "MethodDeclaration":
		a.visitMethod(n)
	
	case "VariableDeclaration":
		a.visitVarDecl(n)
	
	case "Block":
		a.scope = NewScope(a.scope)
		for _, stmt := range n.Statements {
			a.visitNode(stmt)
		}
		a.scope = a.scope.parent
	
	case "Assignment", "AssignmentExpression":
		if len(n.Children) == 2 {
			left := n.Children[0]
			right := n.Children[1]
			a.visitNode(left)
			a.visitNode(right)
		}
	
	case "BinaryOperation", "UnaryOperation":
		for _, c := range n.Children {
			a.visitNode(c)
		}
	
	case "Literal":
		// poxyi
	
	case "Identifier":
		if n.Name != nil {
			if _, ok := a.scope.Resolve(*n.Name); !ok {
				a.addError(fmt.Sprintf("undefined variable '%s'", *n.Name), n.Position)
			}
		}
	
	case "IfStatement", "WhileStatement", "ForStatement":
		for _, c := range n.Children {
			a.visitNode(c)
		}
	
	case "ReturnStatement":
		a.visitReturn(n)
	
	case "MethodCall", "FieldAccess", "ExpressionStatement":
		for _, c := range n.Children {
			a.visitNode(c)
		}
	
	default:
		for _, c := range n.Children {
			a.visitNode(c)
		}
	}
}

func (a *Analyzer) visitClass(n *models.AstNode) {
	if n.Name == nil {
		a.addError("class without name", n.Position)
		return
	}
	
	a.scope = NewScope(a.scope)
	for _, member := range n.Children {
		a.visitNode(member)
	}
	a.scope = a.scope.parent
}

func (a *Analyzer) visitMethod(n *models.AstNode) {
	if n.Name == nil {
		a.addError("method without name", n.Position)
		return
	}
	
	a.scope = NewScope(a.scope)
	prevMethod := a.currentMethod
	a.currentMethod = n
	
	for _, param := range n.Parameters {
		if param.Name != nil {
			if err := a.scope.Define(&Symbol{Name: *param.Name, Type: "param"}); err != nil {
				a.addError(err.Error(), param.Position)
			}
		}
	}
	
	for _, stmt := range n.Statements {
		a.visitNode(stmt)
	}
	
	a.currentMethod = prevMethod
	a.scope = a.scope.parent
}

func (a *Analyzer) visitVarDecl(n *models.AstNode) {
	if n.Name == nil {
		a.addError("variable without name", n.Position)
		return
	}
	name := *n.Name
	if err := a.scope.Define(&Symbol{Name: name, Type: "var"}); err != nil {
		a.addError(err.Error(), n.Position)
	}
}

func (a *Analyzer) visitReturn(n *models.AstNode) {
	if a.currentMethod == nil {
		a.addError("return statement outside of method", n.Position)
		return
	}
	
	hasValue := len(n.Children) > 0
	isVoid := false
	
	if a.currentMethod.ReturnType != nil && a.currentMethod.ReturnType.Name != nil {
		isVoid = *a.currentMethod.ReturnType.Name == "void"
	}
	
	if isVoid && hasValue {
		a.addError("cannot return a value from void method", n.Position)
	}
	
	for _, c := range n.Children {
		a.visitNode(c)
	}
}

func CheckJavaSemantic(data []byte) ([]SemanticError, *models.AstNode, error) {
	var root models.AstNode
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, nil, err
	}
	
	an := NewAnalyzer(&root)
	errs, ok := an.Analyze()
	if len(errs) > 0 {
		return errs, nil, nil
	}
	return nil, ok, nil
}
