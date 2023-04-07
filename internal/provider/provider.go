package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/permitio/permit-golang/pkg/config"
	"github.com/permitio/permit-golang/pkg/permit"
)

// Ensure PermitProvider satisfies various provider interfaces.
var _ provider.Provider = &permitProvider{}

// permitProvider defines the provider implementation.
type permitProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// permitProviderModel describes the provider data model.
type permitProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *permitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "permit"
	resp.Version = p.version
}

func (p *permitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API Key for Permit.io. May also be provided via the PERMIT_API_KEY environment variable.",
				Optional:            true,
			},
		},
		Blocks:      map[string]schema.Block{},
		Description: "Interface with Permit.io",
	}
}

func (p *permitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Permit client")

	// Retrieve provider data from configuration
	var providerConfig permitProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfig)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if providerConfig.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown API Key",
			"The provider cannot create the Permit client as there is an unknown configuration value for the API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PERMIT_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	apiKey := os.Getenv("PERMIT_API_KEY")

	if !providerConfig.ApiKey.IsNull() {
		apiKey = providerConfig.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing API Key",
			"The provider cannot create the Permit client as there is a missing or empty value for the API Key. "+
				"Set the api_key value in the configuration or use the PERMIT_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "permit_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "permit_api_key")

	tflog.Debug(ctx, "Creating Permit client")

	permitConfig := config.NewConfigBuilder(apiKey).Build()

	// Example client configuration for data sources and resources
	client := permit.New(permitConfig)

	// Make the Permit client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Permit client", map[string]any{"success": true})
}

func (p *permitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEnvironmentResource,
		NewProjectResource,
		NewResourceResource,
		NewResourceActionResource,
		NewRoleResource,
		NewTenantResource,
	}
}

func (p *permitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewEnvironmentDataSource,
		NewProjectDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &permitProvider{
			version: version,
		}
	}
}
