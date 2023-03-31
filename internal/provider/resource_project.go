package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/permitio/permit-golang/pkg/models"
	"github.com/permitio/permit-golang/pkg/permit"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectResource{}
var _ resource.ResourceWithImportState = &projectResource{}

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource defines the resource implementation.
type projectResource struct {
	client *permit.Client
}

// projectResourceModel describes the resource data model.
type projectResourceModel struct {
	Key            types.String `tfsdk:"key"`
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Configure adds the provider configured client to the data source.
func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*permit.Client)
	if !ok {
		tflog.Error(ctx, "Unable to prepare client")
		return
	}
	r.client = client

}

func (r *projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project resource",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "Project key",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Project description",
				Optional:            true,
			},
		},
	}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create project resource")

	var plan *projectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new project request")

	projectKey := plan.Key.ValueString()
	projectName := plan.Name.ValueString()
	description := plan.Description.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)
	ctx = tflog.SetField(ctx, "permit_project_name", projectName)

	newProject := *models.NewProjectCreate(projectKey, projectName)

	tflog.Debug(ctx, "Creating project resource")

	if description != "" {
		newProject.SetDescription(description)
	}

	project, err := r.client.Api.Projects.Create(ctx, newProject)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new project request")

	plan.Key = types.StringValue(project.Key)
	plan.Id = types.StringValue(project.Id)
	plan.OrganizationId = types.StringValue(project.OrganizationId)
	plan.Name = types.StringValue(project.Name)
	plan.Description = types.StringValue(*project.Description)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a project")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	tflog.Debug(ctx, "Finished creating project resource", map[string]any{"success": true})
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *projectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *projectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *projectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Api.Projects.Delete(ctx, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Project",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted project resource", map[string]any{"success": true})
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
