package models

type AstNode struct {
	NodeType string         `json:"node_type"`
	Position map[string]int `json:"position"`
	Children []*AstNode     `json:"children,omitempty"`
	
	Name        *string  `json:"name,omitempty"`
	Value       *string  `json:"value,omitempty"`
	LiteralType *string  `json:"literal_type,omitempty"`
	Operator    *string  `json:"operator,omitempty"`
	Modifiers   []string `json:"modifiers,omitempty"`
	IsArray     *bool    `json:"is_array,omitempty"`
	
	ReturnType *AstNode   `json:"return_type,omitempty"`
	FieldType  *AstNode   `json:"field_type,omitempty"`
	Parameters []AstNode  `json:"parameters,omitempty"`
	Arguments  []AstNode  `json:"arguments,omitempty"`
	Statements []*AstNode `json:"statements,omitempty"`
	
	Classes []*AstNode `json:"classes,omitempty"`
	Imports []string   `json:"imports,omitempty"`
}
