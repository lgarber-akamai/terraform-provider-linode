package databasepostgresqlv2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

const (
	DefaultCreateTimeout = 60 * time.Minute
	DefaultUpdateTimeout = 15 * time.Minute
	DefaultDeleteTimeout = 5 * time.Minute
)

func NewResource() resource.Resource {
	return &Resource{
		BaseResource: helper.NewBaseResource(
			helper.BaseResourceConfig{
				Name:   "linode_database_postgresql_v2",
				IDType: types.StringType,
				Schema: &frameworkResourceSchema,
				TimeoutOpts: &timeouts.Opts{
					Update: true,
					Create: true,
					Delete: true,
				},
			},
		),
	}
}

type Resource struct {
	helper.BaseResource
}

func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	tflog.Debug(ctx, "Create linode_database_postgresql_v2")

	var data Model
	client := r.Meta.Client

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, DefaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	createOpts := linodego.PostgresCreateOptions{
		Label:       data.Label.ValueString(),
		Region:      data.Region.ValueString(),
		Type:        data.Type.ValueString(),
		Engine:      data.EngineID.ValueString(),
		ClusterSize: helper.FrameworkSafeInt64ToInt(data.ClusterSize.ValueInt64(), &resp.Diagnostics),
	}

	if !data.AllowList.IsNull() {
		resp.Diagnostics.Append(data.AllowList.ElementsAs(ctx, &createOpts.AllowList, false)...)
	}

	if !data.ForkSource.IsUnknown() && !data.ForkSource.IsNull() {
		createOpts.Fork = &linodego.DatabaseFork{
			Source: helper.FrameworkSafeInt64ToInt(data.ForkSource.ValueInt64(), &resp.Diagnostics),
		}

		if !data.ForkRestoreTime.IsUnknown() && !data.ForkRestoreTime.IsNull() {
			restoreTime, d := data.ForkRestoreTime.ValueRFC3339Time()
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}

			createOpts.Fork.RestoreTime = &restoreTime
		}
	}

	// Handles all errors relevant to create options
	if resp.Diagnostics.HasError() {
		return
	}

	createPoller, err := client.NewEventPollerWithoutEntity(linodego.EntityDatabase, linodego.ActionDatabaseCreate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create event poller",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "client.CreatePostgresDatabase(...)", map[string]any{
		"options": createOpts,
	})

	db, err := client.CreatePostgresDatabase(ctx, createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create PostgreSQL database",
			err.Error(),
		)
		return
	}

	data.Flatten(ctx, db, true)

	createPoller.EntityID = db.ID

	// The `updates` field can only be changed using PUT requests
	if !data.Updates.IsUnknown() && !data.Updates.IsNull() {
		var updates ModelUpdates

		resp.Diagnostics.Append(
			data.Updates.As(
				ctx,
				&updates,
				basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true},
			)...,
		)
		if resp.Diagnostics.HasError() {
			return
		}

		updatesLinodego, d := updates.ToLinodego()
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateOpts := linodego.PostgresUpdateOptions{Updates: &updatesLinodego}

		tflog.Debug(ctx, "client.UpdatePostgresDatabase(...)", map[string]any{
			"options": updateOpts,
		})
		db, err = client.UpdatePostgresDatabase(
			ctx,
			db.ID,
			updateOpts,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update PostgreSQL database",
				err.Error(),
			)
			return
		}
	}

	// We call this twice to ensure the state isn't broken in the case of an update failure
	data.Flatten(ctx, db, true)

	tflog.Debug(ctx, "Waiting for database to finish provisioning")

	// IDs should always be overridden during creation (see #1085)
	// TODO: Remove when Crossplane empty string ID issue is resolved
	data.ID = types.StringValue(strconv.Itoa(db.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if _, err := createPoller.WaitForFinished(ctx, int(createTimeout.Seconds())); err != nil {
		resp.Diagnostics.AddError(
			"Failed to wait for PostgreSQL database to finish creating",
			err.Error(),
		)
	}
}

func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	tflog.Debug(ctx, "Read linode_database_postgresql_v2")

	var data Model
	client := r.Meta.Client

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = populateLogAttributes(ctx, data)

	if helper.FrameworkAttemptRemoveResourceForEmptyID(ctx, data.ID, resp) {
		return
	}

	id := helper.FrameworkSafeStringToInt(data.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := client.GetPostgresDatabase(ctx, id)
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			resp.Diagnostics.AddWarning(
				"Database no longer exists",
				fmt.Sprintf(
					"Removing PostgreSQL database with ID %v from state because it no longer exists",
					id,
				),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to refresh the Database",
			err.Error(),
		)
		return
	}

	data.Flatten(ctx, db, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	tflog.Debug(ctx, "Update linode_database_postgresql_v2")

	client := r.Meta.Client
	var plan, state Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, DefaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	ctx = populateLogAttributes(ctx, state)

	var updateOpts linodego.PostgresUpdateOptions
	shouldUpdate := false

	if !state.Label.Equal(plan.Label) {
		shouldUpdate = true
		updateOpts.Label = plan.Label.ValueString()
	}

	if !state.AllowList.Equal(plan.AllowList) {
		shouldUpdate = true

		var allowList []string

		updateOpts.AllowList = &allowList

		plan.AllowList.ElementsAs(ctx, &allowList, false)
	}

	if !state.Type.Equal(plan.Type) {
		shouldUpdate = true
		updateOpts.Type = plan.Type.ValueString()
	}

	if !state.Updates.Equal(plan.Updates) {
		shouldUpdate = true

		var updates ModelUpdates

		resp.Diagnostics.Append(
			plan.Updates.As(
				ctx,
				&updates,
				basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true},
			)...,
		)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// TODO: Uncomment
	//if !state.EngineID.Equal(plan.EngineID) {
	//	engine, version, err := helper.ParseDatabaseEngineSlug(plan.EngineID.ValueString())
	//	if err != nil {
	//		resp.Diagnostics.AddError("Failed to parse database engine slug", err.Error())
	//		return
	//	}
	//
	//	if engine != state.Engine.ValueString() {
	//		resp.Diagnostics.AddError(
	//			"Cannot update engine component of engine_id",
	//			fmt.Sprintf("%s != %s", engine, state.Engine.ValueString()),
	//		)
	//	}
	//
	//	shouldUpdate = true
	//	updateOpts.Version = version
	//}

	// TODO: Support resizing the cluster

	if shouldUpdate {
		id := helper.FrameworkSafeStringToInt(plan.ID.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "client.UpdatePostgresDatabase(...)", map[string]any{
			"options": updateOpts,
		})
		db, err := client.UpdatePostgresDatabase(ctx, id, updateOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Failed to update database (%d)", id),
				err.Error(),
			)
			return
		}
		plan.Flatten(ctx, db, false)

		// TODO: Poll for updates to complete
	}

	plan.CopyFrom(ctx, &state, true)

	// Workaround for Crossplane issue where ID is not
	// properly populated in plan
	// See TPT-2865 for more details
	if plan.ID.ValueString() == "" {
		plan.ID = state.ID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	tflog.Debug(ctx, "Delete linode_database_postgresql_v2")

	client := r.Meta.Client
	var data Model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, DefaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	ctx = populateLogAttributes(ctx, data)

	id := helper.FrameworkSafeStringToInt(data.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "client.DeletePostgresDatabase(...)")
	err := client.DeletePostgresDatabase(ctx, id)
	if err != nil {
		if lerr, ok := err.(*linodego.Error); (ok && lerr.Code != 404) || !ok {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Failed to delete the database (%s)", data.ID.ValueString()),
				err.Error(),
			)
		}
		return
	}
}

func populateLogAttributes(ctx context.Context, data Model) context.Context {
	return tflog.SetField(ctx, "id", data.ID)
}
