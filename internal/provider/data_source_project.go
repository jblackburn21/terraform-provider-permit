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
var _ datasource.DataSource = &projectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// projectDataSource defines the data source implementation.
type projectDataSource struct {
	client *permit.Client
}

// projectDataSourceModel describes the data source data model.
type projectDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Metadata returns the data source type name.
func (d *projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source.
func (d *projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Project key",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *projectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read project data source")
	var state projectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	projectKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)

	tflog.Debug(ctx, "Reading project data source for key")

	project, err := d.client.Api.Projects.Get(ctx, projectKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updating project data source state")

	// Map project body to model
	state = projectDataSourceModel{
		Id:             types.StringValue(project.Id),
		OrganizationId: types.StringValue(project.OrganizationId),
		Key:            types.StringValue(project.Key),
		Name:           types.StringValue(project.Name),
		Description:    types.StringValue(*project.Description),
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Finished reading project data source", map[string]any{"success": true})
}
