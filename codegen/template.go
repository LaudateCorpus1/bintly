package codegen

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	decodeBaseType = iota
	encodeBaseType
	decodeDerivedBaseType
	encodeDerivedBaseType
	decodeBaseSliceType
	encodeBaseSliceType
	decodeCustomSliceType
	encodeCustomSliceType
	encodeStructType
	decodeStructType
	encodeSliceStructType
	decodeSliceStructType
	encodeEmbeddedAliasTemplate
	decodeEmbeddedAliasSliceTemplate
)

var fieldTemplate = map[int]string{
	encodeBaseType: `	coder.{{.Method}}({{.ReceiverAlias}}.{{.Field}})`,
	decodeBaseType: `	coder.{{.Method}}(&{{.ReceiverAlias}}.{{.Field}})`,
	decodeDerivedBaseType: `	var {{.TransientVar}} {{.BaseType}}
	coder.{{.Method}}(&{{.TransientVar}})
	{{.ReceiverAlias}}.{{.Field}} = {{.FieldType}}({{.TransientVar}})`,
	encodeDerivedBaseType: `	coder.{{.Method}}({{.BaseType}}({{.ReceiverAlias}}.{{.Field}}))`,
	encodeBaseSliceType: `	coder.{{.Method}}(({{.ReceiverAlias}}.{{.Field}}))`,
	decodeBaseSliceType: `	var {{.TransientVar}} []{{.BaseType}}
	coder.{{.Method}}(&{{.TransientVar}})
	{{.ReceiverAlias}}.{{.Field}} = {{.TransientVar}}`,
	encodeCustomSliceType: `	coder.{{.Method}}(*(*[]{{.BaseType}})(unsafe.Pointer(&{{.ReceiverAlias}}.{{.Field}})))`,
	decodeCustomSliceType: `	var {{.TransientVar}} []{{.BaseType}}
	coder.{{.Method}}(&{{.TransientVar}})
	{{.ReceiverAlias}}.{{.Field}} = *(*{{.FieldType}})(unsafe.Pointer(&{{.TransientVar}}))`,
	encodeStructType: `	coder.{{.Method}}({{if .PointerNeeded}}&{{end}}{{.ReceiverAlias}}.{{.Field}})`,
	decodeStructType: `{{if not .PointerNeeded}}	{{.ReceiverAlias}}.{{.Field}} = &{{.FieldType}}{}
{{end}}	coder.{{.Method}}({{if .PointerNeeded}}&{{end}}{{.ReceiverAlias}}.{{.Field}})`,
	encodeSliceStructType: `	var {{.TransientVar}} = len({{.ReceiverAlias}}.{{.Field}})
	coder.Alloc(int32({{.TransientVar}}))
	for i:=0; i < {{.TransientVar}} ; i++ {
		if err := coder.{{.Method}}({{if .PointerNeeded}}&{{end}}{{.ReceiverAlias}}.{{.Field}}[i]);err !=nil {
			return nil
		}
	}`,
	decodeSliceStructType: `	var {{.TransientVar}} = coder.Alloc()
	{{.ReceiverAlias}}.{{.Field}} = make([]{{if not .PointerNeeded}}*{{end}}{{.FieldType}},{{.TransientVar}})
	for i:=0; i < int({{.TransientVar}}) ; i++ {
		if err := coder.{{.Method}}({{if .PointerNeeded}}&{{end}}{{.ReceiverAlias}}.{{.Field}}[i]);err != nil {
			return nil
		}
	}`,
	encodeEmbeddedAliasTemplate: `	if err := coder.Coder(&{{.ReceiverAlias}}.{{.Field}}); err !=nil {
	return err
	}
	`,
	decodeEmbeddedAliasSliceTemplate: `		if err := coder.Coder(&{{.ReceiverAlias}}.{{.Field}}); err != nil {
		return err
	}
	`,
}

const (
	fileCode = iota
	codingStructType
	codingSliceType
)

var blockTemplate = map[int]string{
	fileCode: `// Code generated by bintly codegen. DO NOT EDIT.\n\n

package {{.Pkg}}

import (
{{.Imports}}
)
{{.Code}}

`,
	codingStructType: `
func ({{.Receiver}}) EncodeBinary(coder *bintly.Writer) error {
{{.EncodingCases}}
	return nil
}
func ({{.Receiver}}) DecodeBinary(coder *bintly.Reader) error {
{{.DecodingCases}}	
	return nil
}
`,
	codingSliceType: `func ({{.ReceiverAlias}} *{{.SliceType}}) EncodeBinary(coder *bintly.Writer) error {
	var size = len(*{{.ReceiverAlias}})
	coder.Alloc(int32(size))
	for i:=0; i < size ; i++ {
		if err := coder.Coder(&(*{{.ReceiverAlias}})[i]);err !=nil {
			return nil
		}
	}
	return nil
}

func ({{.ReceiverAlias}} *{{.SliceType}}) DecodeBinary(coder *bintly.Reader) error  {
	var tmp = coder.Alloc()
	*{{.ReceiverAlias}} = make([]{{.ComponentType}},tmp)
	for i:=0; i < int(tmp) ; i++ {
		tmp := 	{{.ComponentType}}{}
		if err := coder.Coder(&(*{{.ReceiverAlias}})[i]);err != nil {
			return nil
		}
		(*{{.ReceiverAlias}})[i] = tmp
	}
	return nil
}
`,
}

func expandTemplate(namespace string, dictionary map[int]string, key int, data interface{}) (string, error) {
	var id = fmt.Sprintf("%v_%v", namespace, key)
	textTemplate, ok := dictionary[key]
	if !ok {
		return "", fmt.Errorf("failed to lookup template for %v.%v", namespace, key)
	}
	temlate, err := template.New(id).Parse(textTemplate)
	if err != nil {
		return "", fmt.Errorf("fiailed to parse template %v %v, due to %v", namespace, key, err)
	}
	writer := new(bytes.Buffer)
	err = temlate.Execute(writer, data)
	return writer.String(), err
}

func expandFieldTemplate(key int, data interface{}) (string, error) {
	return expandTemplate("fieldTemplate", fieldTemplate, key, data)
}

func expandBlockTemplate(key int, data interface{}) (string, error) {
	return expandTemplate("blockTemplate", blockTemplate, key, data)
}
