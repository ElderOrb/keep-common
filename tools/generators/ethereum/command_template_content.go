package main

// commandTemplateContent contains the template string from command.go.tmpl
var commandTemplateContent = `// Code generated - DO NOT EDIT.
// This file is a generated command and any manual changes will be lost.

package cmd

import (
    "sync"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/common/hexutil"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
    "github.com/keep-network/keep-common/pkg/cmd"

    "github.com/keep-network/keep-core/pkg/chain/gen/contract"

    "github.com/urfave/cli"
)

var {{.Class}}Command cli.Command

var {{.FullVar}}Description = ` + "`" + `The {{.DashedName}} command allows calling the {{.Class}} contract on an
	Ethereum network. It has subcommands corresponding to each contract method,
	which respectively each take parameters based on the contract method's
	parameters.

	Subcommands will submit a non-mutating call to the network and output the
	result.

	All subcommands can be called against a specific block by passing the
	-b/--block flag.

	All subcommands can be used to investigate the result of a previous
	transaction that called that same method by passing the -t/--transaction
	flag with the transaction hash.

	Subcommands for mutating methods may be submitted as a mutating transaction
	by passing the -s/--submit flag. In this mode, this command will terminate
	successfully once the transaction has been submitted, but will not wait for
	the transaction to be included in a block. They return the transaction hash.

	Calls that require ether to be paid will get 0 ether by default, which can
	be changed by passing the -v/--value flag.` + "`" + `

func init() {
    AvailableCommands = append(AvailableCommands, cli.Command{
        Name:        "{{.DashedName}}",
        Usage:       ` + "`" + `Provides access to the {{.Class}} contract.` + "`" + `,
        Description: {{.FullVar}}Description,
        Subcommands: []cli.Command{
            {{- $contract := . -}}
            {{- range $i, $method := .ConstMethods }}
            {{- if $method.CommandCallable -}}
                {
                    Name: "{{$method.DashedName}}",
                    Usage: "Calls the {{$method.Modifiers -}} method {{$method.LowerName}} on the {{$contract.Class}} contract.",
                    ArgsUsage: "{{ range $i, $param := $method.ParamInfos -}} [{{$param.Name}}] {{ end }}",
                    Action: {{$contract.ShortVar}}{{$method.CapsName}},
                    Before: cmd.ArgCountChecker({{$method.ParamInfos | len}}),
                    Flags: cmd.ConstFlags,
                },
            {{- end -}}
            {{- end -}}
            {{- range $i, $method := .NonConstMethods }}
            {{- if $method.CommandCallable -}}
                {
                    Name: "{{$method.DashedName}}",
                    Usage: "Calls the {{$method.Modifiers -}} method {{$method.LowerName}} on the {{$contract.Class}} contract.",
                    ArgsUsage: "{{ range $i, $param := $method.ParamInfos -}} [{{$param.Name}}] {{ end }}",
                    Action: {{$contract.ShortVar}}{{$method.CapsName}},
                    Before:
                        {{- if $method.Payable -}}
                        cli.BeforeFunc(cmd.PayableArgsChecker.AndThen(cmd.ArgCountChecker({{$method.ParamInfos | len}})))
                        {{- else -}}
                        cli.BeforeFunc(cmd.NonConstArgsChecker.AndThen(cmd.ArgCountChecker({{$method.ParamInfos | len}})))
                        {{- end }},
                    Flags:
                        {{- if $method.Payable -}}
                        cmd.PayableFlags
                        {{- else -}}
                        cmd.NonConstFlags
                        {{- end }},
                },
            {{- end -}}
            {{- end -}}
        },
    })
}

/// ------------------- Const methods -------------------

{{- $contract := . -}}
{{- range $i, $method := .ConstMethods -}}
{{- if $method.CommandCallable }}

func {{$contract.ShortVar}}{{$method.CapsName}}(c *cli.Context) error {
    contract, err := initialize{{$contract.Class}}(c)
    if err != nil {
        return err
    }

   	{{- range $i, $param := .ParamInfos }}
   	{{$param.Name}}, err := {{$param.ParsingFn}}(c.Args()[{{$i}}])
   	if err != nil {
		return fmt.Errorf(
			"couldn't parse parameter {{$param.Name}}, a {{$param.Type}}, from passed value %v",
			c.Args()[{{$i}}],
		)
   	}
   	{{ end }}

    result, err := contract.{{$method.CapsName}}AtBlock(
        {{$method.Params}}
        cmd.BlockFlagValue.Uint,
    )

    if err != nil {
    	return err
    }

    cmd.PrintOutput(result)

    return nil
}

{{- end -}}
{{- end }}

/// ------------------- Non-const methods -------------------

{{- range $i, $method := .NonConstMethods -}}
{{- if $method.CommandCallable }}

func {{$contract.ShortVar}}{{$method.CapsName}}(c *cli.Context) error {
    contract, err := initialize{{$contract.Class}}(c)
    if err != nil {
        return err
    }

    {{ range $i, $param := .ParamInfos }}
    {{$param.Name}}, err := {{$param.ParsingFn}}(c.Args()[{{$i}}])
    if err != nil {
        return fmt.Errorf(
            "couldn't parse parameter {{$param.Name}}, a {{$param.Type}}, from passed value %v",
            c.Args()[{{$i}}],
        )
    }

    {{ end -}}

    var (
        transaction *types.Transaction
        {{ if gt (len $method.Return.Type) 0 -}}
        result {{$method.Return.Type}}
        {{ end -}}
    )

    if c.Bool(cmd.SubmitFlag) {
        // Do a regular submission. Take payable into account.
        transaction, err = contract.{{$method.CapsName}}(
            {{$method.Params}}
            {{- if $method.Payable -}} cmd.ValueFlagValue.Uint, {{- end -}}
        )
        if err != nil {
            return err
        }

        cmd.PrintOutput(transaction.Hash)
    } else {
        // Do a call.
        {{ if gt (len $method.Return.Type) 0 -}} result, {{ end -}} err = contract.Call{{$method.CapsName}}(
            {{$method.Params}}
            {{- if $method.Payable -}} cmd.ValueFlagValue.Uint, {{- end -}}
            cmd.BlockFlagValue.Uint,
        )
        if err != nil {
            return err
        }

        {{ if gt (len $method.Return.Type) 0 -}}
        cmd.PrintOutput(result)
        {{- else -}}
        cmd.PrintOutput(nil)
        {{- end }}
    }


    return nil
}

{{- end -}}
{{- end }}

/// ------------------- Initialization -------------------

func initialize{{.Class}}(c *cli.Context) (*contract.{{.Class}}, error) {
    config, err := {{.EthereumConfigReader}}(c.GlobalString("config"))
    if err != nil {
        return nil, fmt.Errorf("error reading Ethereum config from file: [%v]", err)
    }

    client, _, _, err := ethutil.ConnectClients(config.URL, config.URLRPC)
    if err != nil {
        return nil, fmt.Errorf("error connecting to Ethereum node: [%v]", err)
    }

    key, err := ethutil.DecryptKeyFile(
        config.Account.KeyFile,
        config.Account.KeyFilePassword,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "failed to read KeyFile: %s: [%v]",
            config.Account.KeyFile,
            err,
        )
    }

    address := common.HexToAddress(config.ContractAddresses["{{.Class}}"])

    return contract.New{{.Class}}(
        address,
        key,
        client,
        &sync.Mutex{},
    )
}
`