package models

type AstNode struct {
	NodeType     string         `json:"node_type"`
	Position     map[string]int `json:"position,omitempty"`
	Name         *string        `json:"name"`
	Value        interface{}    `json:"value"`
	LiteralType  *string        `json:"literal_type"`
	Operator     *string        `json:"operator"`
	Modifiers    []string       `json:"modifiers"`
	IsArray      *bool          `json:"is_array"`
	ReturnType   *AstNode       `json:"return_type"`
	FieldType    *AstNode       `json:"field_type"`
	ParamType    *AstNode       `json:"param_type"`
	Parameters   []AstNode      `json:"parameters"`
	Arguments    []AstNode      `json:"arguments"`
	Statements   []AstNode      `json:"statements"`
	Children     []AstNode      `json:"children"`
	Classes      []AstNode      `json:"classes"`
	Imports      []AstNode      `json:"imports"`
	Fields       []AstNode      `json:"fields"`
	Methods      []AstNode      `json:"methods"`
	GenericTypes []AstNode      `json:"generic_types"`
	Body         *AstNode       `json:"body"`
}
