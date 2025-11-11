package tests

import (
	"encoding/json"
	"testing"
	
	"github.com/romang000/java-semantic-web/internal/models"
	"github.com/romang000/java-semantic-web/internal/service"
	"github.com/stretchr/testify/assert"
)

// helper — выполняет анализ AST и возвращает ошибки
func analyze(t *testing.T, data string) []service.SemanticError {
	var root models.AstNode
	err := json.Unmarshal([]byte(data), &root)
	assert.NoError(t, err, "JSON должен корректно парситься")
	
	an := service.NewAnalyzer(&root)
	errs, _ := an.Analyze()
	return errs
}

func TestUndefinedVariable(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "statements": [
	        {"node_type": "ReturnStatement", "children": [], "position": {"line": 2, "column": 5}}
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "return statement outside of method")
	assert.Equal(t, 2, errs[0].Position["line"])
	assert.Equal(t, 5, errs[0].Position["column"])
}

func TestDuplicateVariable(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Dup",
	      "methods": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "parameters": [],
	          "body": {
	            "node_type": "Block",
	            "statements": [
	              {"node_type": "VariableDeclaration", "name": "x", "position": {"line": 3, "column": 5}},
	              {"node_type": "VariableDeclaration", "name": "x", "position": {"line": 4, "column": 5}}
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "symbol 'x' already defined")
}

func TestReturnValueInVoidMethod(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "methods": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"node_type": "Type", "name": "void"},
	          "parameters": [],
	          "body": {
	            "node_type": "Block",
	            "statements": [
	              {
	                "node_type": "ReturnStatement",
	                "children": [
	                  {"node_type": "Literal", "value": "5", "literal_type": "int", "position": {"line": 5, "column": 12}}
	                ],
	                "position": {"line": 5, "column": 5}
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "cannot return a value from void method")
}

func TestTypeMismatchInExpression(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "methods": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"node_type": "Type", "name": "void"},
	          "parameters": [],
	          "body": {
	            "node_type": "Block",
	            "statements": [
	              {
	                "node_type": "BinaryOperation",
	                "children": [
	                  {"node_type": "Literal", "value": "10", "literal_type": "int"},
	                  {"node_type": "Literal", "value": "hello", "literal_type": "string"}
	                ],
	                "operator": "+"
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "type mismatch")
}

func TestMethodCallArgumentMismatch(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "methods": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "foo",
	          "return_type": {"node_type": "Type", "name": "int"},
	          "parameters": [
	            {"node_type": "Parameter", "name": "a", "param_type": {"node_type": "Type", "name": "int"}}
	          ],
	          "body": {
	            "node_type": "Block",
	            "statements": [
	              {
	                "node_type": "MethodCall",
	                "name": "foo",
	                "arguments": [
	                  {"node_type": "Literal", "value": "hello", "literal_type": "string"}
	                ],
	                "position": {"line": 3, "column": 5}
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "method")
}

func TestFieldAccessUndefined(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "methods": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"node_type": "Type", "name": "void"},
	          "parameters": [],
	          "body": {
	            "node_type": "Block",
	            "statements": [
	              {
	                "node_type": "ExpressionStatement",
	                "children": [
	                  {
	                    "node_type": "FieldAccess",
	                    "children": [
	                      {"node_type": "Identifier", "name": "obj"},
	                      {"node_type": "Identifier", "name": "field"}
	                    ],
	                    "position": {"line": 2, "column": 5}
	                  }
	                ]
	              }
	            ]
	          }
	        }
	      ]
	    }
	  ]
	}`
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "undefined variable 'obj'")
}
