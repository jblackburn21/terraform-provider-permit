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
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &projectResource{}
var _ resource.ResourceWithImportState = &projectResource{}

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

// environmentResource defines the resource implementation.
type environmentResource struct {
	client *permit.Client
}

// environmentResourceModel describes the resource data model.
type environmentResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Configure adds the provider configured client to the data source.
func (r *environmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *environmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Environment resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Environment identifier",
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
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Environment key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Environment name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Environment description",
				Optional:            true,
			},
		},
	}
}

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create environment resource")

	var plan *environmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new environment request")

	projectId := plan.ProjectId.ValueString()
	environmentKey := plan.Key.ValueString()
	environmentName := plan.Name.ValueString()
	environmentDescription := plan.Description.ValueString()

	newEnvironment := *models.NewEnvironmentCreate(environmentKey, environmentName)

	if environmentDescription != "" {
		newEnvironment.SetDescription(environmentDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Setting context for environment")

	r.client.Api.SetContext(ctx, projectId, "")

	tflog.Debug(ctx, "Creating environment resource")

	environment, err := r.client.Api.Environments.Create(ctx, newEnvironment)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create environment",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new environment request")

	plan.Id = types.StringValue(environment.Id)
	plan.OrganizationId = types.StringValue(environment.OrganizationId)
	plan.ProjectId = types.StringValue(environment.ProjectId)
	plan.Key = types.StringValue(environment.Key)
	plan.Name = types.StringValue(environment.Name)
	plan.Description = types.StringValue(*environment.Description)

	tflog.Debug(ctx, "Updating environment state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating environment resource", map[string]any{"success": true})
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read environment resource")

	var state environmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Reading environment resource")

	r.client.Api.SetContext(ctx, projectId, "")

	environment, err := r.client.Api.Environments.Get(ctx, environmentKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read environment",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read environment request")

	// Map response body to model
	state = environmentResourceModel{
		Id:             types.StringValue(environment.Id),
		OrganizationId: types.StringValue(environment.OrganizationId),
		ProjectId:      types.StringValue(environment.ProjectId),
		Key:            types.StringValue(environment.Key),
		Name:           types.StringValue(environment.Name),
		Description:    types.StringValue(*environment.Description),
	}

	tflog.Debug(ctx, "Updating environment state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading environment resource", map[string]any{"success": true})
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update environment resource")

	var plan environmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update environment request")

	projectId := plan.ProjectId.ValueString()
	environmentKey := plan.Key.ValueString()
	environmentName := plan.Name.ValueString()
	environmentDescription := plan.Description.ValueString()

	updateEnvironment := *models.NewEnvironmentUpdate()

	updateEnvironment.SetName(environmentName)

	if environmentDescription != "" {
		updateEnvironment.SetDescription(environmentDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Updating environment resource")

	r.client.Api.SetContext(ctx, projectId, "")

	environment, err := r.client.Api.Environments.Update(ctx, environmentKey, updateEnvironment)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update environment",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update environment request")

	// Overwrite items with refreshed state
	plan = environmentResourceModel{
		Id:             types.StringValue(environment.Id),
		OrganizationId: types.StringValue(environment.OrganizationId),
		ProjectId:      types.StringValue(environment.ProjectId),
		Key:            types.StringValue(environment.Key),
		Name:           types.StringValue(environment.Name),
		Description:    types.StringValue(*environment.Description),
	}

	tflog.Debug(ctx, "Updating environment state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating environment resource", map[string]any{"success": true})
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete environment resource")

	var state *environmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Deleting environment resource")

	r.client.Api.SetContext(ctx, projectId, "")

	err := r.client.Api.Environments.Delete(ctx, environmentKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete environment",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting environment resource", map[string]any{"success": true})
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Preparing to import environment resource")

	split := strings.Split(req.ID, "/")

	if len(split) != 2 {
		resp.Diagnostics.AddError(
			"Error importing environment",
			"Could not import environment, ID should be an {projectId}/{environmentKey}",
		)
		return
	}

	projectKey := split[0]
	environmentKey := split[1]

	project, err := r.client.Api.Projects.Get(ctx, projectKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "permit_project_id", project.Id)
	ctx = tflog.SetField(ctx, "permit_environment_key", environmentKey)

	tflog.Debug(ctx, "Importing environment resource")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), environmentKey)...)
}
