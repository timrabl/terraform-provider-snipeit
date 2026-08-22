// Package acctest is the shared acceptance-test harness for all service
// packages. Acceptance tests live in external test packages (package
// <domain>_test) and use ProtoV6ProviderFactories plus PreCheck.
//
// NOTE: tests inside internal/provider itself (package provider) must NOT
// import this package — that would be an import cycle. They keep local
// equivalents until they migrate into a service package.
package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/timrabl/terraform-provider-snipeit/internal/provider"
)

// ProtoV6ProviderFactories instantiates the provider during acceptance
// testing. The factory function is called for each Terraform CLI command to
// create a provider server the CLI can connect to.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"snipeit": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// PreCheck validates the required environment variables for acceptance tests.
func PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("SNIPEIT_URL") == "" {
		t.Fatal("SNIPEIT_URL must be set for acceptance tests")
	}
	if os.Getenv("SNIPEIT_TOKEN") == "" {
		t.Fatal("SNIPEIT_TOKEN must be set for acceptance tests")
	}
}
