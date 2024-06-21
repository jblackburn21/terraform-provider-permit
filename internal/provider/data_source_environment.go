package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/permitio/permit-golang/pkg/permit"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &environmentDataSource{}

func NewEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

// environmentDataSource defines the data source implementation.
type environmentDataSource struct {
	client *permit.Client
}

// environmentDataSourceModel describes the data source data model.
type environmentDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Metadata returns the data source type name.
func (d *environmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for the data source.
func (d *environmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Environment identifier",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Environment key",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Environment name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Environment description",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *environmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*permit.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *permit.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read environment data source")
	var state environmentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	projectId := state.ProjectId.ValueString()
	environmentKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Setting context for environment")

	d.client.Api.SetContext(ctx, projectId, "")

	tflog.Debug(ctx, "Reading environment data source for key")

	environment, err := d.client.Api.Environments.Get(ctx, environmentKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read environment",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updating environment data source state")

	// Map environment body to model
	state = environmentDataSourceModel{
		Id:             types.StringValue(environment.Id),
		OrganizationId: types.StringValue(environment.OrganizationId),
		ProjectId:      types.StringValue(environment.ProjectId),
		Key:            types.StringValue(environment.Key),
		Name:           types.StringValue(environment.Name),
		Description:    types.StringValue(*environment.Description),
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Finished reading environment data source", map[string]any{"success": true})
}
