package vlan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper"
)

type DataSource struct {
	client *linodego.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	meta := helper.GetDataSourceMeta(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	d.client = meta.Client
}

func (d *DataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "linode_vlans"
}

func (d *DataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = frameworkDatasourceSchema
}

func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data VLANsFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diag := filterConfig.GenerateID(data.Filters)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	data.ID = id

	results, diag := filterConfig.GetAndFilter(
		ctx, d.client,
		data.Filters,
		listVLANs,
		data.Order,
		data.OrderBy,
	)

	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	data.parseVLANs(ctx, helper.AnySliceToTyped[linodego.VLAN](results))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func listVLANs(
	ctx context.Context,
	client *linodego.Client,
	filter string,
) ([]any, error) {
	vlans, err := client.ListVLANs(
		ctx,
		&linodego.ListOptions{Filter: filter},
	)
	if err != nil {
		return nil, err
	}
	return helper.TypedSliceToAny(vlans), nil
}
