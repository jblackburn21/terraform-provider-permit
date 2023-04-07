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
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	Key            types.String `tfsdk:"key"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Project key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
	projectDescription := plan.Description.ValueString()

	newProject := *models.NewProjectCreate(projectKey, projectName)

	if projectDescription != "" {
		newProject.SetDescription(projectDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)

	tflog.Debug(ctx, "Creating project resource")

	project, err := r.client.Api.Projects.Create(ctx, newProject)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new project request")

	plan.Id = types.StringValue(project.Id)
	plan.OrganizationId = types.StringValue(project.OrganizationId)
	plan.Key = types.StringValue(project.Key)
	plan.Name = types.StringValue(project.Name)
	plan.Description = types.StringValue(*project.Description)

	tflog.Debug(ctx, "Updating project state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating project resource", map[string]any{"success": true})
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read project resource")

	var state projectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)

	tflog.Debug(ctx, "Reading project resource")

	project, err := r.client.Api.Projects.Get(ctx, projectKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read project request")

	// Map response body to model
	state = projectResourceModel{
		Id:             types.StringValue(project.Id),
		OrganizationId: types.StringValue(project.OrganizationId),
		Key:            types.StringValue(project.Key),
		Name:           types.StringValue(project.Name),
		Description:    types.StringValue(*project.Description),
	}

	tflog.Debug(ctx, "Updating project state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading project resource", map[string]any{"success": true})
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update project resource")

	var plan projectResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update project request")

	projectKey := plan.Key.ValueString()
	projectName := plan.Name.ValueString()
	projectDescription := plan.Description.ValueString()

	updateProject := *models.NewProjectUpdate()

	updateProject.SetName(projectName)

	if projectDescription != "" {
		updateProject.SetDescription(projectDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)

	tflog.Debug(ctx, "Updating project resource")

	project, err := r.client.Api.Projects.Update(ctx, projectKey, updateProject)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update project request")

	// Overwrite items with refreshed state
	plan = projectResourceModel{
		Id:             types.StringValue(project.Id),
		OrganizationId: types.StringValue(project.OrganizationId),
		Key:            types.StringValue(project.Key),
		Name:           types.StringValue(project.Name),
		Description:    types.StringValue(*project.Description),
	}

	tflog.Debug(ctx, "Updating project state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating project resource", map[string]any{"success": true})
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete project resource")

	var state *projectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_key", projectKey)

	tflog.Debug(ctx, "Deleting project resource")

	err := r.client.Api.Projects.Delete(ctx, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting project resource", map[string]any{"success": true})
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
