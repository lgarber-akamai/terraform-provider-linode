//go:build integration || consumerimagesharegrouptoken

package consumerimagesharegrouptoken_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/linode/terraform-provider-linode/v3/linode/acceptance"
	"github.com/linode/terraform-provider-linode/v3/linode/consumerimagesharegrouptoken/tmpl"
)

// This test requires two separate Linode API tokens, one for the producer
// and one for the consumer.
//
// These can be set using the LINODE_PRODUCER_TOKEN and LINODE_CONSUMER_TOKEN
// environment variables.
//
// If either is not set,the test will be skipped.
func TestAccResourceImageShareGroupToken_basic(t *testing.T) {
	t.Parallel()

	if os.Getenv("LINODE_PRODUCER_TOKEN") == "" || os.Getenv("LINODE_CONSUMER_TOKEN") == "" {
		t.Skip("Skipping test: both LINODE_PRODUCER_TOKEN and LINODE_CONSUMER_TOKEN must be set")
	}

	// Define provider factories for both producer and consumer
	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"linode-producer": acceptance.NewTokenProviderFactory("LINODE_PRODUCER_TOKEN"),
		"linode-consumer": acceptance.NewTokenProviderFactory("LINODE_CONSUMER_TOKEN"),
	}

	resourceName := "linode_consumer_image_share_group_token.foobar"
	shareGroupLabel := acctest.RandomWithPrefix("tf-test")
	tokenLabel := acctest.RandomWithPrefix("tf-test")
	tokenLabelUpdated := tokenLabel + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, shareGroupLabel, tokenLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "token_uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "label", tokenLabel),
					resource.TestCheckResourceAttrSet(resourceName, "valid_for_sharegroup_uuid"),
					resource.TestCheckNoResourceAttr(resourceName, "sharegroup_uuid"),
					resource.TestCheckNoResourceAttr(resourceName, "sharegroup_label"),
				),
			},
			{
				Config: tmpl.Basic(t, shareGroupLabel, tokenLabelUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "token_uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "label", tokenLabelUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "valid_for_sharegroup_uuid"),
					resource.TestCheckNoResourceAttr(resourceName, "sharegroup_uuid"),
					resource.TestCheckNoResourceAttr(resourceName, "sharegroup_label"),
				),
			},
		},
	})
}
