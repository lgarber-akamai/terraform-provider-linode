package databasepostgresqlv2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

type ModelFork struct {
	Source types.Int64 `tfsdk:"source"`
}

type ModelForkDetails struct {
	RestoreTime timetypes.RFC3339 `tfsdk:"restore_time"`
}

type ModelHosts struct {
	Primary   types.String `tfsdk:"primary"`
	Secondary types.String `tfsdk:"secondary"`
}

type ModelUpdates struct {
	DayOfWeek types.Int64  `tfsdk:"day_of_week"`
	Duration  types.Int64  `tfsdk:"duration"`
	Frequency types.String `tfsdk:"frequency"`
	HourOfDay types.Int64  `tfsdk:"hour_of_day"`
}

func (m ModelUpdates) ToLinodego() (linodego.DatabaseMaintenanceWindow, diag.Diagnostics) {
	var d diag.Diagnostics

	return linodego.DatabaseMaintenanceWindow{
		DayOfWeek: linodego.DatabaseDayOfWeek(helper.FrameworkSafeInt64ToInt(m.DayOfWeek.ValueInt64(), &d)),
		Duration:  helper.FrameworkSafeInt64ToInt(m.Duration.ValueInt64(), &d),
		Frequency: linodego.DatabaseMaintenanceFrequency(m.Frequency.ValueString()),
		HourOfDay: helper.FrameworkSafeInt64ToInt(m.HourOfDay.ValueInt64(), &d),
	}, d
}

type ModelPendingUpdate struct {
	Deadline    timetypes.RFC3339 `tfsdk:"deadline"`
	Description types.String      `tfsdk:"description"`
	PlannedFor  timetypes.RFC3339 `tfsdk:"planned_for"`
}

type Model struct {
	Timeouts timeouts.Value `tfsdk:"timeouts"`

	ID types.String `tfsdk:"id"`

	Label    types.String `tfsdk:"label"`
	EngineID types.String `tfsdk:"engine_id"`
	Region   types.String `tfsdk:"region"`
	Type     types.String `tfsdk:"type"`

	AllowList     types.Set   `tfsdk:"allow_list"`
	ClusterSize   types.Int64 `tfsdk:"cluster_size"`
	SSLConnection types.Bool  `tfsdk:"ssl_connection"`

	Created           timetypes.RFC3339 `tfsdk:"created"`
	Encrypted         types.Bool        `tfsdk:"encrypted"`
	Engine            types.String      `tfsdk:"engine"`
	Members           types.Map         `tfsdk:"members"`
	OldestRestoreTime timetypes.RFC3339 `tfsdk:"oldest_restore_time"`
	Platform          types.String      `tfsdk:"platform"`
	Port              types.Int64       `tfsdk:"port"`
	Status            types.String      `tfsdk:"status"`
	Updated           timetypes.RFC3339 `tfsdk:"updated"`
	Version           types.String      `tfsdk:"version"`

	Hosts types.Object `tfsdk:"hosts"`

	Fork        types.Object `tfsdk:"fork"`
	ForkDetails types.Object `tfsdk:"fork_details"`

	Updates        types.Object `tfsdk:"updates"`
	PendingUpdates types.Set    `tfsdk:"pending_updates"`
}

func (m *Model) Flatten(ctx context.Context, db *linodego.PostgresDatabase, preserveKnown bool) (d diag.Diagnostics) {
	m.ID = helper.KeepOrUpdateString(m.ID, strconv.Itoa(db.ID), preserveKnown)

	m.Label = helper.KeepOrUpdateString(m.Label, db.Label, preserveKnown)
	m.EngineID = helper.KeepOrUpdateString(m.EngineID, fmt.Sprintf("%s/%s", db.Engine, strings.Split(db.Version, ".")[0]), preserveKnown)
	m.Region = helper.KeepOrUpdateString(m.Region, db.Region, preserveKnown)
	m.Type = helper.KeepOrUpdateString(m.Type, db.Type, preserveKnown)
	m.AllowList = helper.KeepOrUpdateSet(
		types.StringType,
		m.AllowList,
		helper.StringSliceToFrameworkValueSlice(db.AllowList),
		preserveKnown,
		&d,
	)
	if d.HasError() {
		return
	}
	m.ClusterSize = helper.KeepOrUpdateInt64(m.ClusterSize, int64(db.ClusterSize), preserveKnown)
	m.SSLConnection = helper.KeepOrUpdateBool(m.SSLConnection, db.SSLConnection, preserveKnown)
	m.Created = helper.KeepOrUpdateValue(m.Created, timetypes.NewRFC3339TimePointerValue(db.Created), preserveKnown)
	m.Encrypted = helper.KeepOrUpdateBool(m.Encrypted, db.Encrypted, preserveKnown)
	m.Engine = helper.KeepOrUpdateString(m.Engine, db.Engine, preserveKnown)

	membersCasted := helper.MapMap(
		db.Members,
		func(key string, value linodego.DatabaseMemberType) (string, string) {
			return key, string(value)
		},
	)

	m.Members = helper.KeepOrUpdateStringMap(ctx, m.Members, membersCasted, preserveKnown, &d)
	if d.HasError() {
		return
	}

	m.OldestRestoreTime = helper.KeepOrUpdateValue(m.OldestRestoreTime, timetypes.NewRFC3339TimePointerValue(db.OldestRestoreTime), preserveKnown)
	m.Platform = helper.KeepOrUpdateString(m.Platform, string(db.Platform), preserveKnown)
	m.Port = helper.KeepOrUpdateInt64(m.Port, int64(db.Port), preserveKnown)
	m.Status = helper.KeepOrUpdateString(m.Status, string(db.Status), preserveKnown)
	m.Updated = helper.KeepOrUpdateValue(m.Updated, timetypes.NewRFC3339TimePointerValue(db.Updated), preserveKnown)
	m.Version = helper.KeepOrUpdateString(m.Version, db.Version, preserveKnown)

	hostsObject, rd := types.ObjectValueFrom(
		ctx,
		hostsAttributes,
		&ModelHosts{
			Primary:   types.StringValue(db.Hosts.Primary),
			Secondary: types.StringValue(db.Hosts.Secondary),
		},
	)
	d.Append(rd...)
	m.Hosts = helper.KeepOrUpdateValue(m.Hosts, hostsObject, preserveKnown)

	forkObject, rd := types.ObjectValueFrom(
		ctx,
		forkAttributes,
		&ModelFork{
			Source: types.Int64Value(int64(db.Fork.Source)),
		},
	)
	d.Append(rd...)
	m.Fork = helper.KeepOrUpdateValue(m.Fork, forkObject, preserveKnown)

	forkDetailsObject, rd := types.ObjectValueFrom(
		ctx,
		forkDetailsAttributes,
		&ModelForkDetails{
			RestoreTime: timetypes.NewRFC3339TimePointerValue(db.Fork.RestoreTime),
		},
	)
	d.Append(rd...)
	m.ForkDetails = helper.KeepOrUpdateValue(m.ForkDetails, forkDetailsObject, preserveKnown)

	updatesObject, rd := types.ObjectValueFrom(
		ctx,
		updatesAttributes,
		&ModelUpdates{
			DayOfWeek: types.Int64Value(int64(db.Updates.DayOfWeek)),
			Duration:  types.Int64Value(int64(db.Updates.Duration)),
			Frequency: types.StringValue(string(db.Updates.Frequency)),
			HourOfDay: types.Int64Value(int64(db.Updates.HourOfDay)),
		},
	)
	d.Append(rd...)
	m.Updates = helper.KeepOrUpdateValue(m.Updates, updatesObject, preserveKnown)

	pendingObjects := helper.MapSlice(
		db.Updates.Pending,
		func(pending linodego.DatabaseMaintenanceWindowPending) types.Object {
			result, rd := types.ObjectValueFrom(
				ctx,
				pendingUpdateAttributes,
				&ModelPendingUpdate{
					Deadline:    timetypes.NewRFC3339TimePointerValue(pending.Deadline),
					Description: types.StringValue(pending.Description),
					PlannedFor:  timetypes.NewRFC3339TimePointerValue(pending.PlannedFor),
				},
			)
			d.Append(rd...)

			return result
		},
	)

	pendingSet, rd := types.SetValueFrom(
		ctx,
		types.ObjectType{
			AttrTypes: pendingUpdateAttributes,
		},
		pendingObjects,
	)
	d.Append(rd...)

	m.PendingUpdates = helper.KeepOrUpdateValue(m.PendingUpdates, pendingSet, preserveKnown)

	return nil
}

func (m *Model) CopyFrom(ctx context.Context, other *Model, preserveKnown bool) {
	m.ID = helper.KeepOrUpdateValue(m.ID, other.ID, preserveKnown)
	m.Engine = helper.KeepOrUpdateValue(m.Engine, other.Engine, preserveKnown)
	m.EngineID = helper.KeepOrUpdateValue(m.EngineID, other.EngineID, preserveKnown)
	m.Label = helper.KeepOrUpdateValue(m.Label, other.Label, preserveKnown)
	m.Region = helper.KeepOrUpdateValue(m.Region, other.Region, preserveKnown)
	m.Type = helper.KeepOrUpdateValue(m.Type, other.Type, preserveKnown)
	m.AllowList = helper.KeepOrUpdateValue(m.AllowList, other.AllowList, preserveKnown)
	m.ClusterSize = helper.KeepOrUpdateValue(m.ClusterSize, other.ClusterSize, preserveKnown)
	m.SSLConnection = helper.KeepOrUpdateValue(m.SSLConnection, other.SSLConnection, preserveKnown)
	m.Created = helper.KeepOrUpdateValue(m.Created, other.Created, preserveKnown)
	m.Encrypted = helper.KeepOrUpdateValue(m.Encrypted, other.Encrypted, preserveKnown)
	m.Members = helper.KeepOrUpdateValue(m.Members, other.Members, preserveKnown)
	m.OldestRestoreTime = helper.KeepOrUpdateValue(m.OldestRestoreTime, other.OldestRestoreTime, preserveKnown)
	m.Platform = helper.KeepOrUpdateValue(m.Platform, other.Platform, preserveKnown)
	m.Port = helper.KeepOrUpdateValue(m.Port, other.Port, preserveKnown)
	m.Status = helper.KeepOrUpdateValue(m.Status, other.Status, preserveKnown)
	m.Updated = helper.KeepOrUpdateValue(m.Updated, other.Updated, preserveKnown)
	m.Version = helper.KeepOrUpdateValue(m.Version, other.Version, preserveKnown)
	m.Hosts = helper.KeepOrUpdateValue(m.Hosts, other.Hosts, preserveKnown)
	m.Fork = helper.KeepOrUpdateValue(m.Fork, other.Fork, preserveKnown)
	m.ForkDetails = helper.KeepOrUpdateValue(m.ForkDetails, other.ForkDetails, preserveKnown)
	m.Updates = helper.KeepOrUpdateValue(m.Updates, other.Updates, preserveKnown)
	m.PendingUpdates = helper.KeepOrUpdateValue(m.PendingUpdates, other.PendingUpdates, preserveKnown)
}
