{{- /* gotype: morpheus/pkg/fsm-generator.TemplateModel */ -}}
package {{ .Model.Name | snake }}

// Code generated by fsm-generator. YOU SHOULD EDIT THIS FILE

// Service is an interface which is used for business-logic encapsulation
// provide list of methods for calling in state handlers
type Service interface {
}