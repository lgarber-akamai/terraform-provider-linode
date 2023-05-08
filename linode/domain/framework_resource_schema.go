package domain

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const domainSecondsDescription = "Valid values are 0, 30, 120, 300, 3600, 7200, 14400, 28800, " +
	"57600, 86400, 172800, 345600, 604800, 1209600, and 2419200 - any other value will be rounded to " +
	"the nearest valid value."

var frameworkResourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The Domain's unique ID.",
			Computed:    true,
		},

		"domain": schema.StringAttribute{
			Description: "The domain this Domain represents. These must be unique in our system; you cannot have " +
				"two Domains representing the same domain.",
			Required: true,
		},
		"type": schema.StringAttribute{
			Description: "If this Domain represents the authoritative source of information for the domain it " +
				"describes, or if it is a read-only copy of a master (also called a slave).",
			Required: true,
			Default:  stringdefault.StaticString("master"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOfCaseInsensitive("master", "slave"),
			},
		},
		"group": schema.StringAttribute{
			Description: "The group this Domain belongs to. This is for display purposes only.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(0, 50),
			},
		},
		"status": schema.StringAttribute{
			Description: "Used to control whether this Domain is currently being rendered.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("active"),
		},
		"description": schema.StringAttribute{
			Description: "A description for this Domain. This is for display purposes only.",
			Validators: []validator.String{
				stringvalidator.LengthBetween(0, 253),
			},
			Optional: true,
		},
		"master_ips": schema.SetAttribute{
			Description: "The IP addresses representing the master DNS for this Domain.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"axfr_ips": schema.SetAttribute{
			Description: "The list of IPs that may perform a zone transfer for this Domain. This is potentially " +
				"dangerous, and should be set to an empty list unless you intend to use it.",
			Optional: true,
		},
		"ttl_sec": schema.Int64Attribute{
			Description: "'Time to Live' - the amount of time in seconds that this Domain's records may be " +
				"cached by resolvers or other domain servers. " + domainSecondsDescription,
			Optional: true,
			// TODO: implement rounding
		},
		"retry_sec": schema.Int64Attribute{
			Description: "The interval, in seconds, at which a failed refresh should be retried. " +
				domainSecondsDescription,
			Optional: true,
			// TODO: implementing rounding
		},
		"expire_sec": schema.Int64Attribute{
			Description: "The amount of time in seconds that may pass before this Domain is no longer " +
				domainSecondsDescription,
			Optional: true,
			// TODO: implement rounding
		},
		"refresh_sec": schema.Int64Attribute{
			Description: "The amount of time in seconds before this Domain should be refreshed. " +
				domainSecondsDescription,
			Optional: true,
			// TODO: implement rounding
		},
		"soa_email": schema.StringAttribute{
			Description: "Start of Authority email address. This is required for master Domains.",
			Optional:    true,
		},
		"tags": schema.SetAttribute{
			Description: "An array of tags applied to this object. Tags are for organizational purposes only.",
			Optional:    true,
			ElementType: types.StringType,
		},
	},
}
