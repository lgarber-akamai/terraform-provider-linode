package vpcsubnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper"
	"github.com/linode/terraform-provider-linode/linode/helper/customtypes"
)

type VPCSubnetModel struct {
	ID      types.Int64                        `tfsdk:"id"`
	VPCId   types.Int64                        `tfsdk:"vpc_id"`
	Label   types.String                       `tfsdk:"label"`
	IPv4    types.String                       `tfsdk:"ipv4"`
	Linodes types.List                         `tfsdk:"linodes"`
	Created customtypes.RFC3339TimeStringValue `tfsdk:"created"`
	Updated customtypes.RFC3339TimeStringValue `tfsdk:"updated"`
}

func (d *VPCSubnetModel) parseComputedAttributes(
	ctx context.Context,
	subnet *linodego.VPCSubnet,
) diag.Diagnostics {
	d.ID = types.Int64Value(int64(subnet.ID))

	linodes, diag := types.ListValueFrom(ctx, types.Int64Type, subnet.Linodes)
	if diag != nil {
		return diag
	}
	d.Linodes = linodes

	d.Created = customtypes.RFC3339TimeStringValue{
		StringValue: helper.NullableTimeToFramework(subnet.Created),
	}
	d.Updated = customtypes.RFC3339TimeStringValue{
		StringValue: helper.NullableTimeToFramework(subnet.Updated),
	}

	return nil
}

func (d *VPCSubnetModel) parseVPCSubnet(
	ctx context.Context,
	subnet *linodego.VPCSubnet,
) diag.Diagnostics {
	d.Label = types.StringValue(subnet.Label)
	d.IPv4 = types.StringValue(subnet.IPv4)

	return d.parseComputedAttributes(ctx, subnet)
}