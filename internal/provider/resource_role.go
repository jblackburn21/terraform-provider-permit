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
var _ resource.Resource = &roleResource{}
var _ resource.ResourceWithImportState = &roleResource{}

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

// roleResource defines the resource implementation.
type roleResource struct {
	client *permit.Client
}

// tenantResourceModel describes the resource data model.
type roleResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	EnvironmentId  types.String `tfsdk:"environment_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
}

// Configure adds the provider configured client to the data source.
func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Role resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Role identifier",
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
				MarkdownDescription: "Role key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Role name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Role description",
				Optional:            true,
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "Role permissions",
				ElementType:         types.StringType,
				Required:            true,
			},
		},
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create role resource")

	var plan *roleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new role request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	roleKey := plan.Key.ValueString()
	roleName := plan.Name.ValueString()
	roleDescription := plan.Description.ValueString()

	var rolePermissions []string

	resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &rolePermissions, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	newRole := *models.NewRoleCreate(roleKey, roleName)

	if roleDescription != "" {
		newRole.SetDescription(roleDescription)
	}

	newRole.SetPermissions(rolePermissions)

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_role_key", roleKey)

	tflog.Debug(ctx, "Setting context for role")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tflog.Debug(ctx, "Creating role resource")

	role, err := r.client.Api.Roles.Create(ctx, newRole)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create role",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new role request")

	permissions, _ := types.SetValueFrom(ctx, types.StringType, role.Permissions)

	plan.Id = types.StringValue(role.Id)
	plan.OrganizationId = types.StringValue(role.OrganizationId)
	plan.ProjectId = types.StringValue(role.ProjectId)
	plan.EnvironmentId = types.StringValue(role.EnvironmentId)
	plan.Key = types.StringValue(role.Key)
	plan.Name = types.StringValue(role.Name)
	plan.Description = types.StringValue(*role.Description)
	plan.Permissions = permissions

	tflog.Debug(ctx, "Updating role state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating role resource", map[string]any{"success": true})
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read role resource")

	var state roleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	roleKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_role_key", roleKey)

	tflog.Debug(ctx, "Reading role resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	role, err := r.client.Api.Roles.Get(ctx, roleKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read role",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read role request")

	// Map response body to model
	state = roleResourceModel{
		Id:             types.StringValue(role.Id),
		OrganizationId: types.StringValue(role.OrganizationId),
		ProjectId:      types.StringValue(role.ProjectId),
		EnvironmentId:  types.StringValue(role.EnvironmentId),
		Key:            types.StringValue(role.Key),
		Name:           types.StringValue(role.Name),
		Description:    types.StringValue(*role.Description),
	}

	tflog.Debug(ctx, "Updating role state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading role resource", map[string]any{"success": true})
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update role resource")

	var plan roleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update role request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	roleKey := plan.Key.ValueString()
	roleName := plan.Name.ValueString()
	roleDescription := plan.Description.ValueString()

	updateRole := *models.NewRoleUpdate()

	updateRole.SetName(roleName)

	if roleDescription != "" {
		updateRole.SetDescription(roleDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_role_key", roleKey)

	tflog.Debug(ctx, "Updating role resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	role, err := r.client.Api.Roles.Update(ctx, roleKey, updateRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update role",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update role request")

	// Overwrite items with refreshed state
	plan = roleResourceModel{
		Id:             types.StringValue(role.Id),
		OrganizationId: types.StringValue(role.OrganizationId),
		ProjectId:      types.StringValue(role.ProjectId),
		EnvironmentId:  types.StringValue(role.EnvironmentId),
		Key:            types.StringValue(role.Key),
		Name:           types.StringValue(role.Name),
		Description:    types.StringValue(*role.Description),
	}

	tflog.Debug(ctx, "Updating role state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating role resource", map[string]any{"success": true})
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete role resource")

	var state *roleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	roleKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_role_key", roleKey)

	tflog.Debug(ctx, "Deleting role resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	err := r.client.Api.Roles.Delete(ctx, roleKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete role",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting role resource", map[string]any{"success": true})
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Preparing to import role resource")

	split := strings.Split(req.ID, "/")

	if len(split) != 3 {
		resp.Diagnostics.AddError(
			"Error importing role",
			"Could not import role, ID should be an {project-key}/{environment-key}/{role-key}",
		)
		return
	}

	projectKey := split[0]
	environmentKey := split[1]
	roleKey := split[2]

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
	ctx = tflog.SetField(ctx, "permit_role_key", roleKey)

	tflog.Debug(ctx, "Importing tenant resource")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environment.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), roleKey)...)
}
