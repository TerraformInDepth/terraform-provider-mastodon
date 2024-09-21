// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mastodon": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	client_host := os.Getenv("MASTODON_HOST")
	assert.NotEmpty(t, client_host, "MASTODON_HOST must be set for acceptance tests")

	client_id := os.Getenv("MASTODON_CLIENT_ID")
	assert.NotEmpty(t, client_id, "MASTODON_CLIENT_ID must be set for acceptance tests")

	client_secret := os.Getenv("MASTODON_CLIENT_SECRET")
	assert.NotEmpty(t, client_secret, "MASTODON_CLIENT_SECRET must be set for acceptance tests")

	client_token := os.Getenv("MASTODON_ACCESS_TOKEN")
	assert.NotEmpty(t, client_token, "MASTODON_ACCESS_TOKEN must be set for acceptance tests")
}
