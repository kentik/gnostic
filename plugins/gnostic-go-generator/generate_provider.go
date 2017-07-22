package main

import "strings"

func (renderer *ServiceRenderer) GenerateProvider() ([]byte, error) {
	f := NewLineWriter()
	f.WriteLine("// GENERATED FILE: DO NOT EDIT!\n")
	f.WriteLine("package " + renderer.Model.Package)
	f.WriteLine(``)
	f.WriteLine(`// To create a server, first write a class that implements this interface.`)
	f.WriteLine(`// Then pass an instance of it to Initialize().`)
	f.WriteLine(`type Provider interface {`)
	for _, method := range renderer.Model.Methods {
		f.WriteLine(``)
		f.WriteLine(`// Provider`)
		f.WriteLine(commentForText(method.Description))
		if method.hasParameters() {
			if method.hasResponses() {
				f.WriteLine(method.ProcessorName +
					`(parameters *` + method.ParametersTypeName +
					`, responses *` + method.ResponsesTypeName + `) (err error)`)
			} else {
				f.WriteLine(method.ProcessorName + `(parameters *` + method.ParametersTypeName + `) (err error)`)
			}
		} else {
			if method.hasResponses() {
				f.WriteLine(method.ProcessorName + `(responses *` + method.ResponsesTypeName + `) (err error)`)
			} else {
				f.WriteLine(method.ProcessorName + `() (err error)`)
			}
		}
	}
	f.WriteLine(`}`)
	return f.Bytes(), nil
}

func commentForText(text string) string {
	result := ""
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += "// " + line
	}
	return result
}