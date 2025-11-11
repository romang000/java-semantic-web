package service

import (
	"bytes"
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
	Name   string
	Type   string   // тип переменной или метода
	Params []string // для метода: типы параметров
	Return string   // для метода: тип возвращаемого значения
}

type Scope struct {
	symbols map[string]*Symbol
	parent  *Scope
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

type Analyzer struct {
	root          *models.AstNode
	scope         *Scope
	errors        []SemanticError
	currentMethod *models.AstNode
	inFieldAccess bool
}

func NewAnalyzer(root *models.AstNode) *Analyzer {
	return &Analyzer{
		root:  root,
		scope: NewScope(nil),
	}
}

func (a *Analyzer) addError(message string, position map[string]int) {
	a.errors = append(a.errors, SemanticError{Message: message, Position: position})
}

func (a *Analyzer) Analyze() ([]SemanticError, *models.AstNode) {
	a.visitNode(a.root)
	if len(a.errors) > 0 {
		return a.errors, nil
	}
	return nil, a.root
}

func (a *Analyzer) visitNode(n *models.AstNode) {
	if n == nil {
		return
	}
	
	switch n.NodeType {
	case "Program":
		for i := range n.Classes {
			a.visitNode(&n.Classes[i])
		}
	case "ClassDeclaration":
		a.visitClass(n)
	case "MethodDeclaration":
		a.visitMethod(n)
	case "VariableDeclaration", "FieldDeclaration":
		a.visitVarDecl(n)
	case "Block":
		a.scope = NewScope(a.scope)
		for i := range n.Statements {
			a.visitNode(&n.Statements[i])
		}
		a.scope = a.scope.parent
	case "Assignment", "AssignmentExpression":
		if len(n.Children) == 2 {
			leftType := a.getExprType(&n.Children[0])
			rightType := a.getExprType(&n.Children[1])
			if leftType != rightType && leftType != "unknown" && rightType != "unknown" {
				a.addError(fmt.Sprintf("type mismatch in assignment: %s vs %s", leftType, rightType), n.Position)
			}
			a.visitNode(&n.Children[0])
			a.visitNode(&n.Children[1])
		}
	case "UnaryOperation", "IfStatement", "WhileStatement", "ForStatement", "ExpressionStatement":
		for i := range n.Children {
			a.visitNode(&n.Children[i])
		}
		for i := range n.Arguments {
			a.visitNode(&n.Arguments[i])
		}
	
	case "BinaryOperation":
		a.visitNode(&n.Children[0])
		a.visitNode(&n.Children[1])
		
		leftType := a.getExprType(&n.Children[0])
		rightType := a.getExprType(&n.Children[1])
		
		if leftType != rightType {
			a.addError(
				fmt.Sprintf("type mismatch: %s %s %s", leftType, *n.Operator, rightType),
				n.Position,
			)
		}
	case "Identifier":
		if n.Name != nil {
			// Проверяем только если мы не находимся в FieldAccess
			if !a.inFieldAccess {
				if _, ok := a.scope.Resolve(*n.Name); !ok {
					a.addError(fmt.Sprintf("undefined variable '%s'", *n.Name), n.Position)
				}
			}
		}
	
	case "FieldAccess":
		prev := a.inFieldAccess
		a.inFieldAccess = true
		if len(n.Children) > 0 {
			// проверяем левый-most элемент (объект)
			left := &n.Children[0]
			if left.NodeType == "Identifier" {
				if _, ok := a.scope.Resolve(*left.Name); !ok {
					a.addError(fmt.Sprintf("undefined variable '%s'", *left.Name), left.Position)
				}
			} else {
				a.visitNode(left)
			}
			
			// остальные элементы цепочки просто обходим
			for i := 1; i < len(n.Children); i++ {
				a.visitNode(&n.Children[i])
			}
		}
		a.inFieldAccess = prev
	case "MethodCall":
		a.visitMethodCall(n)
	case "ReturnStatement":
		a.visitReturn(n)
	default:
		for i := range n.Children {
			a.visitNode(&n.Children[i])
		}
		for i := range n.Arguments {
			a.visitNode(&n.Arguments[i])
		}
	}
}

func (a *Analyzer) visitClass(n *models.AstNode) {
	if n.Name == nil {
		a.addError("class without name", n.Position)
		return
	}
	a.scope = NewScope(a.scope)
	for i := range n.Children {
		a.visitNode(&n.Children[i])
	}
	for i := range n.Fields {
		a.visitVarDecl(&n.Fields[i])
	}
	for i := range n.Methods {
		a.visitMethod(&n.Methods[i])
	}
	for i := range n.Statements {
		a.visitNode(&n.Statements[i])
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
	
	paramTypes := []string{}
	for i := range n.Parameters {
		param := &n.Parameters[i]
		if param.Name != nil {
			t := "unknown"
			if param.ParamType != nil && param.ParamType.Name != nil {
				t = *param.ParamType.Name
			}
			paramTypes = append(paramTypes, t)
			if err := a.scope.Define(&Symbol{Name: *param.Name, Type: t}); err != nil {
				a.addError(err.Error(), param.Position)
			}
		}
	}
	
	retType := "void"
	if n.ReturnType != nil && n.ReturnType.Name != nil {
		retType = *n.ReturnType.Name
	}
	
	// Зарегистрируем метод в текущей области
	if n.Name != nil {
		a.scope.Define(&Symbol{Name: *n.Name, Type: "method", Params: paramTypes, Return: retType})
	}
	
	if n.Body != nil {
		a.visitNode(n.Body)
	}
	
	for i := range n.Statements {
		a.visitNode(&n.Statements[i])
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
	typ := "unknown"
	if n.FieldType != nil && n.FieldType.Name != nil {
		typ = *n.FieldType.Name
	}
	if err := a.scope.Define(&Symbol{Name: name, Type: typ}); err != nil {
		a.addError(err.Error(), n.Position)
	}
}

func (a *Analyzer) visitReturn(n *models.AstNode) {
	if a.currentMethod == nil {
		a.addError("return statement outside of method", n.Position)
		return
	}
	
	expectedType := "void"
	if a.currentMethod.ReturnType != nil && a.currentMethod.ReturnType.Name != nil {
		expectedType = *a.currentMethod.ReturnType.Name
	}
	
	if len(n.Children) > 0 {
		retType := a.getExprType(&n.Children[0])
		if expectedType != "void" && retType != expectedType && retType != "unknown" {
			a.addError(fmt.Sprintf("return type mismatch: expected %s, got %s", expectedType, retType), n.Position)
		}
		if expectedType == "void" {
			a.addError("cannot return a value from void method", n.Position)
		}
	} else if expectedType != "void" {
		a.addError(fmt.Sprintf("method must return a value of type %s", expectedType), n.Position)
	}
	
	for i := range n.Children {
		a.visitNode(&n.Children[i])
	}
}

// Вычисление типа выражения
func (a *Analyzer) getExprType(n *models.AstNode) string {
	if n == nil {
		return "unknown"
	}
	
	switch n.NodeType {
	case "Literal":
		if n.LiteralType != nil {
			return *n.LiteralType
		}
	case "Identifier":
		if sym, ok := a.scope.Resolve(*n.Name); ok {
			return sym.Type
		}
	case "BinaryOperation":
		if len(n.Children) == 2 {
			left := a.getExprType(&n.Children[0])
			right := a.getExprType(&n.Children[1])
			if left != right && left != "unknown" && right != "unknown" {
				a.addError(fmt.Sprintf("type mismatch in binary operation: %s vs %s", left, right), n.Position)
				return "error"
			}
			return left
		}
	case "MethodCall":
		return a.handleMethodCall(n)
	}
	
	return "unknown"
}

// Новый метод для обработки MethodCall
func (a *Analyzer) handleMethodCall(n *models.AstNode) string {
	if n == nil {
		return "unknown"
	}
	
	var sym *Symbol
	
	// Находим метод по имени
	if n.Name != nil {
		s, ok := a.scope.Resolve(*n.Name)
		if ok && s.Type == "method" {
			sym = s
		} else {
			a.addError(fmt.Sprintf("undefined method '%s'", *n.Name), n.Position)
		}
	}
	
	for i := range n.Children {
		a.visitNode(&n.Children[i])
	}
	for i := range n.Arguments {
		a.visitNode(&n.Arguments[i])
	}
	
	if sym != nil {
		if len(n.Arguments) != len(sym.Params) {
			a.addError(fmt.Sprintf("method call argument count mismatch: expected %d, got %d", len(sym.Params), len(n.Arguments)), n.Position)
		} else {
			for i := range n.Arguments {
				argType := a.getExprType(&n.Arguments[i])
				if argType != sym.Params[i] && argType != "unknown" {
					a.addError(fmt.Sprintf("method call argument type mismatch: expected %s, got %s", sym.Params[i], argType), n.Arguments[i].Position)
				}
			}
		}
		return sym.Return
	}
	
	return "unknown"
}

func (a *Analyzer) visitMethodCall(n *models.AstNode) {
	if n == nil {
		return
	}
	
	for i := range n.Children {
		a.visitNode(&n.Children[i])
	}
	for i := range n.Arguments {
		a.visitNode(&n.Arguments[i])
	}
	
	a.getExprType(n)
}

func CheckJavaSemantic(data []byte) ([]SemanticError, *models.AstNode, error) {
	clean := bytes.ReplaceAll(data, []byte{0}, []byte{})
	
	if len(clean) == 0 {
		return nil, nil, fmt.Errorf("empty JSON after cleaning")
	}
	
	var root models.AstNode
	if err := json.Unmarshal(clean, &root); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	an := NewAnalyzer(&root)
	errs, ok := an.Analyze()
	if len(errs) > 0 {
		return errs, nil, nil
	}
	return nil, ok, nil
}
