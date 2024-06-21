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
var _ resource.Resource = &tenantResource{}
var _ resource.ResourceWithImportState = &tenantResource{}

func NewTenantResource() resource.Resource {
	return &tenantResource{}
}

// tenantResource defines the resource implementation.
type tenantResource struct {
	client *permit.Client
}

// tenantResourceModel describes the resource data model.
type tenantResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	ProjectId      types.String `tfsdk:"project_id"`
	EnvironmentId  types.String `tfsdk:"environment_id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

// Configure adds the provider configured client to the data source.
func (r *tenantResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *tenantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *tenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Tenant resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Tenant identifier",
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
				MarkdownDescription: "Tenant key",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Tenant name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Tenant description",
				Optional:            true,
			},
		},
	}
}

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create tenant resource")

	var plan *tenantResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building new tenant request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	tenantKey := plan.Key.ValueString()
	tenantName := plan.Name.ValueString()
	tenantDescription := plan.Description.ValueString()

	newTenant := *models.NewTenantCreate(tenantKey, tenantName)

	if tenantDescription != "" {
		newTenant.SetDescription(tenantDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_tenant_key", tenantKey)

	tflog.Debug(ctx, "Setting context for tenant")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tflog.Debug(ctx, "Creating tenant resource")

	tenant, err := r.client.Api.Tenants.Create(ctx, newTenant)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create tenant",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed new tenant request")

	plan.Id = types.StringValue(tenant.Id)
	plan.OrganizationId = types.StringValue(tenant.OrganizationId)
	plan.ProjectId = types.StringValue(tenant.ProjectId)
	plan.EnvironmentId = types.StringValue(tenant.EnvironmentId)
	plan.Key = types.StringValue(tenant.Key)
	plan.Name = types.StringValue(tenant.Name)
	plan.Description = types.StringValue(*tenant.Description)

	tflog.Debug(ctx, "Updating tenant state")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished creating tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read tenant resource")

	var state tenantResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	tenantKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_tenant_key", tenantKey)

	tflog.Debug(ctx, "Reading resource resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tenant, err := r.client.Api.Tenants.Get(ctx, tenantKey)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read tenant",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed read tenant request")

	// Map response body to model
	state = tenantResourceModel{
		Id:             types.StringValue(tenant.GetId()),
		OrganizationId: types.StringValue(tenant.GetOrganizationId()),
		ProjectId:      types.StringValue(tenant.GetProjectId()),
		EnvironmentId:  types.StringValue(tenant.GetEnvironmentId()),
		Key:            types.StringValue(tenant.GetKey()),
		Name:           types.StringValue(tenant.GetName()),
		Description:    types.StringValue(tenant.GetDescription()),
	}

	tflog.Debug(ctx, "Updating tenant state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished reading tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update tenant resource")

	var plan tenantResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Building update tenant request")

	projectId := plan.ProjectId.ValueString()
	environmentId := plan.EnvironmentId.ValueString()
	tenantKey := plan.Key.ValueString()
	tenantName := plan.Name.ValueString()
	tenantDescription := plan.Description.ValueString()

	updateTenant := *models.NewTenantUpdate()

	updateTenant.SetName(tenantName)

	if tenantDescription != "" {
		updateTenant.SetDescription(tenantDescription)
	}

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_tenant_key", tenantKey)

	tflog.Debug(ctx, "Updating tenant resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	tenant, err := r.client.Api.Tenants.Update(ctx, tenantKey, updateTenant)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update tenant",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Completed update tenant request")

	// Overwrite items with refreshed state
	plan = tenantResourceModel{
		Id:             types.StringValue(tenant.GetId()),
		OrganizationId: types.StringValue(tenant.GetOrganizationId()),
		ProjectId:      types.StringValue(tenant.GetProjectId()),
		EnvironmentId:  types.StringValue(tenant.GetEnvironmentId()),
		Key:            types.StringValue(tenant.GetKey()),
		Name:           types.StringValue(tenant.GetName()),
		Description:    types.StringValue(tenant.GetDescription()),
	}

	tflog.Debug(ctx, "Updating tenant state")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Finished updating tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete tenant resource")

	var state *tenantResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := state.ProjectId.ValueString()
	environmentId := state.EnvironmentId.ValueString()
	tenantKey := state.Key.ValueString()

	ctx = tflog.SetField(ctx, "permit_project_id", projectId)
	ctx = tflog.SetField(ctx, "permit_environment_id", environmentId)
	ctx = tflog.SetField(ctx, "permit_tenant_key", tenantKey)

	tflog.Debug(ctx, "Deleting tenant resource")

	r.client.Api.SetContext(ctx, projectId, environmentId)

	err := r.client.Api.Tenants.Delete(ctx, tenantKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete tenant",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Finished deleting tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Preparing to import tenant resource")

	split := strings.Split(req.ID, "/")

	if len(split) != 3 {
		resp.Diagnostics.AddError(
			"Error importing tenant",
			"Could not import tenant, ID should be an {project-key}/{environment-key}/{tenant-key}",
		)
		return
	}

	projectKey := split[0]
	environmentKey := split[1]
	tenantKey := split[2]

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
	ctx = tflog.SetField(ctx, "permit_tenant_key", tenantKey)

	tflog.Debug(ctx, "Importing tenant resource")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), project.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), environment.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), tenantKey)...)
}
