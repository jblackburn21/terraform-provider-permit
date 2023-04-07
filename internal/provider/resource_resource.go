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
var _ resource.Resource = &resourceResource{}
var _ resource.ResourceWithImportState = &resourceResource{}

func NewResourceResource() resource.Resource {
	return &resourceResource{}
}

// resourceResource defines the resource implementation.
type resourceResource struct {
	client *permit.Client
}

// resourceResourceModel describes the resource data model.
type resourceResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	EnvironmentId  types.String `tfsdk:"environment_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	// TODO: attributes
}

// Configure adds the provider configured client to the data source.
func (r *resourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *resourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *resourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier",
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
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Environment identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Resource key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Resource description",
				Optional:            true,
			},
		},
	}
}

func (r *resourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create resource resource")

	var plan *resourceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new resource request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	resourceKey := plan.Key.ValueString()
	resourceName := plan.Name.ValueString()
	resourceDescription := plan.Description.ValueString()

	resourceActions := make(map[string]models.ActionBlockEditable)

	newResource := *models.NewResourceCreate(resourceKey, resourceName, resourceActions)

	if resourceDescription != "" {
		newResource.SetDescription(resourceDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_key", resourceKey)

	tflog.Debug(ctx, "Setting context for resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tflog.Debug(ctx, "Creating resource resource")

	permitResource, err := r.client.Api.Resources.Create(ctx, newResource)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create resource",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new resource request")

	plan.Id = types.StringValue(permitResource.Id)
	plan.OrganizationId = types.StringValue(permitResource.OrganizationId)
	plan.ProjectId = types.StringValue(permitResource.ProjectId)
	plan.EnvironmentId = types.StringValue(permitResource.EnvironmentId)
	plan.Key = types.StringValue(permitResource.Key)
	plan.Name = types.StringValue(permitResource.Name)
	plan.Description = types.StringValue(*permitResource.Description)

	tflog.Debug(ctx, "Updating resource state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating resource resource", map[string]any{"success": true})
}

func (r *resourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read resource resource")

	var state resourceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	resourceKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_key", resourceKey)

	tflog.Debug(ctx, "Reading resource resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	permitResource, err := r.client.Api.Resources.Get(ctx, resourceKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read resource",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read resource request")

	// Map response body to model
	state = resourceResourceModel{
		Id:             types.StringValue(permitResource.Id),
		OrganizationId: types.StringValue(permitResource.OrganizationId),
		ProjectId:      types.StringValue(permitResource.ProjectId),
		EnvironmentId:  types.StringValue(permitResource.EnvironmentId),
		Key:            types.StringValue(permitResource.Key),
		Name:           types.StringValue(permitResource.Name),
		Description:    types.StringValue(*permitResource.Description),
	}

	tflog.Debug(ctx, "Updating resource state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading resource resource", map[string]any{"success": true})
}

func (r *resourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update resource resource")

	var plan resourceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update resource request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	resourceKey := plan.Key.ValueString()
	resourceName := plan.Name.ValueString()
	resourceDescription := plan.Description.ValueString()

	updateResource := *models.NewResourceUpdate()

	updateResource.SetName(resourceName)

	if resourceDescription != "" {
		updateResource.SetDescription(resourceDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_key", resourceKey)

	tflog.Debug(ctx, "Updating resource resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	permitResource, err := r.client.Api.Resources.Update(ctx, resourceKey, updateResource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update resource",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update resource request")

	// Overwrite items with refreshed state
	plan = resourceResourceModel{
		Id:             types.StringValue(permitResource.Id),
		OrganizationId: types.StringValue(permitResource.OrganizationId),
		ProjectId:      types.StringValue(permitResource.ProjectId),
		EnvironmentId:  types.StringValue(permitResource.EnvironmentId),
		Key:            types.StringValue(permitResource.Key),
		Name:           types.StringValue(permitResource.Name),
		Description:    types.StringValue(*permitResource.Description),
	}

	tflog.Debug(ctx, "Updating resource state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating resource resource", map[string]any{"success": true})
}

func (r *resourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete resource resource")

	var state *resourceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	resourceKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_key", resourceKey)

	tflog.Debug(ctx, "Deleting resource resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	err := r.client.Api.Resources.Delete(ctx, resourceKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete resource",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting resource resource", map[string]any{"success": true})
}

func (r *resourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Preparing to import resource resource")

	split := strings.Split(req.ID, "/")

	if len(split) != 3 {
		resp.Diagnostics.AddError(
			"Error importing resource",
			"Could not import resource, ID should be an {project-key}/{environment-key}/{resource-key}",
		)
		return
	}

	projectKey := split[0]
	environmentKey := split[1]
	resourceKey := split[2]

	project, err := r.client.Api.Projects.Get(ctx, projectKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read project",
			err.Error(),
		)
		return
	}

	r.client.Api.SetContext(ctx, project.Id, "")

	environment, err := r.client.Api.Environments.Get(ctx, environmentKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read environment",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "permit_project_id", project.Id)
	ctx = tflog.SetField(ctx, "permit_environment_id", environment.Id)
	ctx = tflog.SetField(ctx, "permit_resource_key", resourceKey)

	tflog.Debug(ctx, "Importing resource resource")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environment.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), resourceKey)...)
}
