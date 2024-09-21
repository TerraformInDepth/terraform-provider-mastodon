// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = IdentityFunction{}
)

func NewIdentityFunction() function.Function {
	return IdentityFunction{}
}

type IdentityFunction struct{}

func (r IdentityFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "identity"
}

func (r IdentityFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "Identity function",
		MarkdownDescription: "Returns the identity string of a user from a specific server.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "username",
				MarkdownDescription: "The username to generate the identity from.",
			},
			function.StringParameter{
				Name:                "server",
				MarkdownDescription: "The server the user is hosted on.",
			},
		},
		Return: function.StringReturn{},
	}
}

func (r IdentityFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var username string
	var server string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &username, &server))

	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, "@"+username+"@"+server))
}
