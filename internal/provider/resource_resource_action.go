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
var _ resource.Resource = &resourceActionResource{}
var _ resource.ResourceWithImportState = &resourceActionResource{}

func NewResourceActionResource() resource.Resource {
	return &resourceActionResource{}
}

// resourceActionResource defines the resource implementation.
type resourceActionResource struct {
	client *permit.Client
}

// resourceActionResourceModel describes the resource data model.
type resourceActionResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	EnvironmentId  types.String `tfsdk:"environment_id"`
	ResourceId     types.String `tfsdk:"resource_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Configure adds the provider configured client to the data source.
func (r *resourceActionResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *resourceActionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_action"
}

func (r *resourceActionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource action resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource action identifier",
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
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Resource action key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource action name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Resource action description",
				Optional:            true,
			},
		},
	}
}

func (r *resourceActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create resource action resource")

	var plan *resourceActionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new resource action request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	resourceId := plan.ResourceId.ValueString()
	resourceActionKey := plan.Key.ValueString()
	resourceActionName := plan.Name.ValueString()
	resourceActionDescription := plan.Description.ValueString()

	newResourceAction := *models.NewResourceActionCreate(resourceActionKey, resourceActionName)

	if resourceActionDescription != "" {
		newResourceAction.SetDescription(resourceActionDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_id", resourceId)
	ctx = tflog.SetField(ctx, "permit_resource_action_key", resourceActionKey)

	tflog.Debug(ctx, "Setting context for resource action")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tflog.Debug(ctx, "Creating resource action resource")

	resourceAction, err := r.client.Api.ResourceActions.Create(ctx, resourceId, newResourceAction)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create resource action",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new resource action request")

	plan.Id = types.StringValue(resourceAction.Id)
	plan.OrganizationId = types.StringValue(resourceAction.OrganizationId)
	plan.ProjectId = types.StringValue(resourceAction.ProjectId)
	plan.EnvironmentId = types.StringValue(resourceAction.EnvironmentId)
	plan.ResourceId = types.StringValue(resourceAction.ResourceId)
	plan.Key = types.StringValue(resourceAction.Key)
	plan.Name = types.StringValue(resourceAction.Name)
	plan.Description = types.StringValue(*resourceAction.Description)

	tflog.Debug(ctx, "Updating resource action state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating resource action resource", map[string]any{"success": true})
}

func (r *resourceActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read resource action resource")

	var state resourceActionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	resourceId := state.ResourceId.ValueString()
	resourceActionKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_id", resourceId)
	ctx = tflog.SetField(ctx, "permit_resource_action_key", resourceActionKey)

	tflog.Debug(ctx, "Reading resource action resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	resourceAction, err := r.client.Api.ResourceActions.Get(ctx, resourceId, resourceActionKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read resource action",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read resource action request")

	// Map response body to model
	state = resourceActionResourceModel{
		Id:             types.StringValue(resourceAction.Id),
		OrganizationId: types.StringValue(resourceAction.OrganizationId),
		ProjectId:      types.StringValue(resourceAction.ProjectId),
		EnvironmentId:  types.StringValue(resourceAction.EnvironmentId),
		ResourceId:     types.StringValue(resourceAction.ResourceId),
		Key:            types.StringValue(resourceAction.Key),
		Name:           types.StringValue(resourceAction.Name),
		Description:    types.StringValue(*resourceAction.Description),
	}

	tflog.Debug(ctx, "Updating resource action state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading resource action resource", map[string]any{"success": true})
}

func (r *resourceActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update resource action resource")

	var plan resourceActionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update resource action request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	resourceId := plan.ResourceId.ValueString()
	resourceActionKey := plan.Key.ValueString()
	resourceActionName := plan.Name.ValueString()
	resourceActionDescription := plan.Description.ValueString()

	updateResourceAction := *models.NewResourceActionUpdate()

	updateResourceAction.SetName(resourceActionName)

	if resourceActionDescription != "" {
		updateResourceAction.SetDescription(resourceActionDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_id", resourceId)
	ctx = tflog.SetField(ctx, "permit_resource_action_key", resourceActionKey)

	tflog.Debug(ctx, "Updating resource action resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	resourceAction, err := r.client.Api.ResourceActions.Update(ctx, resourceId, resourceActionKey, updateResourceAction)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update resource action",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update resource action request")

	// Overwrite items with refreshed state
	plan = resourceActionResourceModel{
		Id:             types.StringValue(resourceAction.Id),
		OrganizationId: types.StringValue(resourceAction.OrganizationId),
		ProjectId:      types.StringValue(resourceAction.ProjectId),
		EnvironmentId:  types.StringValue(resourceAction.EnvironmentId),
		ResourceId:     types.StringValue(resourceAction.ResourceId),
		Key:            types.StringValue(resourceAction.Key),
		Name:           types.StringValue(resourceAction.Name),
		Description:    types.StringValue(*resourceAction.Description),
	}

	tflog.Debug(ctx, "Updating resource action state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating resource action resource", map[string]any{"success": true})
}

func (r *resourceActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete resource action resource")

	var state *resourceActionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	resourceId := state.ResourceId.ValueString()
	resourceActionKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_resource_id", resourceId)
	ctx = tflog.SetField(ctx, "permit_resource_action_key", resourceActionKey)

	tflog.Debug(ctx, "Deleting resource action resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	err := r.client.Api.ResourceActions.Delete(ctx, resourceId, resourceActionKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete resource action",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting resource action resource", map[string]any{"success": true})
}

func (r *resourceActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Preparing to import resource action resource")

	split := strings.Split(req.ID, "/")

	if len(split) != 4 {
		resp.Diagnostics.AddError(
			"Error importing resource action",
			"Could not import resource action, ID should be an {project-key}/{environment-key}/{resource-key}/{resource-action-key}",
		)
		return
	}

	projectKey := split[0]
	environmentKey := split[1]
	resourceKey := split[2]
	resourceActionKey := split[3]

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

	r.client.Api.SetContext(ctx, project.Id, environment.Id)

	permitResource, err := r.client.Api.Resources.Get(ctx, resourceKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read resource",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "permit_project_id", project.Id)
	ctx = tflog.SetField(ctx, "permit_environment_id", environment.Id)
	ctx = tflog.SetField(ctx, "permit_resource_id", permitResource.Id)
	ctx = tflog.SetField(ctx, "permit_resource_action_key", resourceActionKey)

	tflog.Debug(ctx, "Import resource action resource")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environment.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_id"), permitResource.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), resourceActionKey)...)
}
