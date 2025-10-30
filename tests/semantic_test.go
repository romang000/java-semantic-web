package tests

import (
	"encoding/json"
	"github.com/romang000/java-semantic-web/internal/service"
	"testing"
	
	"github.com/romang000/java-semantic-web/internal/models"
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

func TestSemanticAnalyzer_ValidProgram(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "position": {"line": 1, "column": 1},
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "HelloWorld",
	      "children": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"name": "void"},
	          "parameters": [
	            {"node_type": "Parameter", "name": "args"}
	          ],
	          "statements": [
	            {"node_type": "VariableDeclaration", "name": "x"},
	            {"node_type": "Assignment",
	              "children": [
	                {"node_type": "Identifier", "name": "x"},
	                {"node_type": "Literal", "value": "10"}
	              ]
	            }
	          ]
	        }
	      ]
	    }
	  ]
	}`
	
	errs := analyze(t, data)
	assert.Len(t, errs, 0, "корректный код не должен выдавать ошибок")
}

func TestSemanticAnalyzer_UndefinedVariable(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "children": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "foo",
	          "return_type": {"name": "void"},
	          "statements": [
	            {"node_type": "Identifier", "name": "x", "position": {"line": 2, "column": 5}}
	          ]
	        }
	      ]
	    }
	  ]
	}`
	
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "undefined variable 'x'")
	assert.Equal(t, 2, errs[0].Position["line"])
	assert.Equal(t, 5, errs[0].Position["column"])
}

func TestSemanticAnalyzer_DuplicateVariable(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Dup",
	      "children": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"name": "void"},
	          "statements": [
	            {"node_type": "VariableDeclaration", "name": "x", "position": {"line": 3, "column": 5}},
	            {"node_type": "VariableDeclaration", "name": "x", "position": {"line": 4, "column": 5}}
	          ]
	        }
	      ]
	    }
	  ]
	}`
	
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "symbol 'x' already defined")
}

func TestSemanticAnalyzer_ReturnOutsideMethod(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Broken",
	      "children": [
	        {"node_type": "ReturnStatement", "position": {"line": 3, "column": 3}}
	      ]
	    }
	  ]
	}`
	
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "return statement outside of method")
}

func TestSemanticAnalyzer_ReturnValueInVoidMethod(t *testing.T) {
	data := `{
	  "node_type": "Program",
	  "classes": [
	    {
	      "node_type": "ClassDeclaration",
	      "name": "Test",
	      "children": [
	        {
	          "node_type": "MethodDeclaration",
	          "name": "main",
	          "return_type": {"name": "void"},
	          "statements": [
	            {
	              "node_type": "ReturnStatement",
	              "position": {"line": 5, "column": 5},
	              "children": [
	                {"node_type": "Literal", "value": "5", "position": {"line": 5, "column": 12}}
	              ]
	            }
	          ]
	        }
	      ]
	    }
	  ]
	}`
	
	errs := analyze(t, data)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "cannot return a value from void method")
}
