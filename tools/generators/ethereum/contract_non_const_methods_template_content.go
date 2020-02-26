package main

// contractNonConstMethodsTemplateContent contains the template string from contract_non_const_methods.go.tmpl
var contractNonConstMethodsTemplateContent = `{{- $contract := . -}}
{{- $logger := (print $contract.ShortVar "Logger") -}}
{{- range $i, $method := .NonConstMethods }}

// Transaction submission.
func ({{$contract.ShortVar}} *{{$contract.Class}}) {{$method.CapsName}}(
	{{$method.ParamDeclarations -}}
	{{- if $method.Payable -}}
	value *big.Int,
	{{ end }}
	transactionOptions ...ethutil.TransactionOptions,
) (*types.Transaction, error) {
	{{$logger}}.Debug(
		"submitting transaction {{$method.LowerName}}\n",
		{{$method.Params}}
		{{- if $method.Payable -}}
		"\nValue: ", value,
		{{ end -}}
	)

	{{$contract.ShortVar}}.transactionMutex.Lock()
	defer {{$contract.ShortVar}}.transactionMutex.Unlock()

	// create a copy
    transactorOptions := &(*{{$contract.ShortVar}}.transactorOptions)

    {{if $method.Payable -}}
    transactorOptions.Value = value
    {{- end }}

	if len(transactionOptions) > 1 {
		return nil, fmt.Errorf(
			"could not process multiple transaction options sets",
		)
	} else if len(transactionOptions) > 0 {
		transactionOptions[0].Apply(transactorOptions)
	}

	transaction, err := {{$contract.ShortVar}}.contract.{{$method.CapsName}}(
		transactorOptions,
		{{$method.Params}}
	)

	if err != nil {
		return transaction, {{$contract.ShortVar}}.errorResolver.ResolveError(
			err,
			{{$contract.ShortVar}}.transactorOptions.From,
			{{if $method.Payable -}}
			value
			{{- else -}}
			nil
			{{- end -}},
			"{{$method.LowerName}}",
			{{$method.Params}}
		)
	}

	{{$logger}}.Debugf(
		"submitted transaction {{$method.LowerName}} with id: [%v]",
		transaction.Hash().Hex(),
	)

	return transaction, err
}

{{- $returnVar := print "result, " -}}
{{ if eq $method.Return.Type "" -}}
{{- $returnVar = "" -}}
{{- end }}

// Non-mutating call, not a transaction submission.
func ({{$contract.ShortVar}} *{{$contract.Class}}) Call{{$method.CapsName}}(
	{{$method.ParamDeclarations -}}
	{{- if $method.Payable -}}
	value *big.Int,
	{{ end -}}
	blockNumber *big.Int,
) ({{- if gt (len $method.Return.Type) 0 -}} {{$method.Return.Type}}, {{- end -}} error) {
	{{- if gt (len $method.Return.Type) 0 }}
	var result {{$method.Return.Type}}
	{{- else }}
	var result interface{} = nil
	{{- end }}

	err := ethutil.CallAtBlock(
		{{$contract.ShortVar}}.transactorOptions.From,
		blockNumber,
		{{- if $method.Payable -}}
		value,
		{{ else -}}
		nil,
		{{ end -}}
		{{$contract.ShortVar}}.contractABI,
		{{$contract.ShortVar}}.caller,
		{{$contract.ShortVar}}.errorResolver,
		{{$contract.ShortVar}}.contractAddress,
		"{{$method.LowerName}}",
		&result,
		{{$method.Params}}
	)

	return {{$returnVar}}err
}

{{- end -}}
`
