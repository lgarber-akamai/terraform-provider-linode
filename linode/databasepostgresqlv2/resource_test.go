//go:build integration || databasepostgresqlv2

package databasepostgresqlv2_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/linode/terraform-provider-linode/v2/linode/helper"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/acceptance"
	"github.com/linode/terraform-provider-linode/v2/linode/databasepostgresqlv2/tmpl"
)

var testRegion, testEngine string

func init() {
	resource.AddTestSweepers("linode_database_postgresql_v2", &resource.Sweeper{
		Name: "linode_database_postgresql_v2",
		F:    sweep,
	})

	client, err := acceptance.GetTestClient()
	if err != nil {
		log.Fatal(err)
	}

	region, err := acceptance.GetRandomRegionWithCaps([]string{"Managed Databases"}, "core")
	if err != nil {
		log.Fatal(err)
	}

	testRegion = region

	engine, err := helper.ResolveValidDBEngine(
		context.Background(),
		*client,
		string(linodego.DatabaseEngineTypePostgres),
	)
	if err != nil {
		log.Fatal(err)
	}

	testEngine = engine.ID
}

func sweep(prefix string) error {
	client, err := acceptance.GetTestClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	listOpts := acceptance.SweeperListOptions(prefix, "label")

	dbs, err := client.ListPostgresDatabases(context.Background(), listOpts)
	if err != nil {
		return fmt.Errorf("error getting postgres databases: %w", err)
	}
	for _, db := range dbs {
		if !acceptance.ShouldSweep(prefix, db.Label) {
			continue
		}
		err := client.DeletePostgresDatabase(context.Background(), db.ID)
		if err != nil {
			return fmt.Errorf("error destroying %s during sweep: %w", db.Label, err)
		}
	}

	return nil
}

func TestAccResource_basic(t *testing.T) {
	t.Parallel()

	resName := "linode_database_postgresql_v2.foobar"
	label := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, label, testRegion, testEngine, "g6-nanode-1"),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckPostgresDatabaseExists(resName, nil),

					resource.TestCheckResourceAttrSet(resName, "id"),
					resource.TestCheckResourceAttr(resName, "label", label),
					resource.TestCheckResourceAttr(resName, "engine_id", testEngine),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "cluster_size", "1"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "true"),
					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttr(resName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resName, "engine", "postgresql"),
					resource.TestCheckResourceAttrSet(resName, "members.%"),
					resource.TestCheckNoResourceAttr(resName, "oldest_restore_time"),
					resource.TestCheckResourceAttr(resName, "platform", "rdbms-default"),
					resource.TestCheckResourceAttrSet(resName, "port"),
					// resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "version"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "1"),
					resource.TestCheckResourceAttr(resName, "allow_list.0", "0.0.0.0/0"),

					resource.TestCheckResourceAttrSet(resName, "hosts.primary"),
					resource.TestCheckResourceAttr(resName, "hosts.secondary", ""),

					resource.TestCheckNoResourceAttr(resName, "fork_source"),
					resource.TestCheckNoResourceAttr(resName, "fork_restore_time"),

					resource.TestCheckResourceAttrSet(resName, "updates.day_of_week"),
					resource.TestCheckResourceAttrSet(resName, "updates.duration"),
					resource.TestCheckResourceAttrSet(resName, "updates.frequency"),
					resource.TestCheckResourceAttrSet(resName, "updates.hour_of_day"),

					resource.TestCheckResourceAttr(resName, "pending_updates.#", "0"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResource_complex(t *testing.T) {
	t.Parallel()

	resName := "linode_database_postgresql_v2.foobar"
	label := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV5ProviderFactories: acceptance.ProtoV5ProviderFactories,
		CheckDestroy:             acceptance.CheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Complex(
					t,
					tmpl.TemplateData{
						Label:       label,
						Region:      testRegion,
						EngineID:    testEngine,
						Type:        "g6-nanode-1",
						AllowedIP:   "10.0.0.3/32",
						ClusterSize: 1,
						Updates: tmpl.TemplateDataUpdates{
							HourOfDay: 3,
							DayOfWeek: 2,
							Duration:  4,
							Frequency: "weekly",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckPostgresDatabaseExists(resName, nil),

					resource.TestCheckResourceAttrSet(resName, "id"),
					resource.TestCheckResourceAttr(resName, "label", label),
					resource.TestCheckResourceAttr(resName, "engine_id", testEngine),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "cluster_size", "1"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "true"),
					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttr(resName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resName, "engine", "postgresql"),
					resource.TestCheckResourceAttrSet(resName, "members.%"),
					resource.TestCheckResourceAttr(resName, "platform", "rdbms-default"),
					resource.TestCheckResourceAttrSet(resName, "port"),
					resource.TestCheckNoResourceAttr(resName, "oldest_restore_time"),
					// resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "version"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "1"),
					resource.TestCheckResourceAttr(resName, "allow_list.0", "10.0.0.3/32"),

					resource.TestCheckResourceAttrSet(resName, "hosts.primary"),

					resource.TestCheckNoResourceAttr(resName, "fork_source"),
					resource.TestCheckNoResourceAttr(resName, "fork_restore_time"),

					resource.TestCheckResourceAttr(resName, "updates.day_of_week", "2"),
					resource.TestCheckResourceAttr(resName, "updates.duration", "4"),
					resource.TestCheckResourceAttr(resName, "updates.frequency", "weekly"),
					resource.TestCheckResourceAttr(resName, "updates.hour_of_day", "3"),

					resource.TestCheckResourceAttr(resName, "pending_updates.#", "0"),
				),
			},
			{
				Config: tmpl.Complex(
					t,
					tmpl.TemplateData{
						Label:       label,
						Region:      testRegion,
						EngineID:    testEngine,
						Type:        "g6-standard-1",
						AllowedIP:   "10.0.0.4/32",
						ClusterSize: 3,
						Updates: tmpl.TemplateDataUpdates{
							HourOfDay: 2,
							DayOfWeek: 3,
							Duration:  4,
							Frequency: "weekly",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					acceptance.CheckPostgresDatabaseExists(resName, nil),

					resource.TestCheckResourceAttrSet(resName, "id"),
					resource.TestCheckResourceAttr(resName, "label", label),
					resource.TestCheckResourceAttr(resName, "engine_id", testEngine),
					resource.TestCheckResourceAttr(resName, "region", testRegion),
					resource.TestCheckResourceAttr(resName, "type", "g6-standard-1"),
					resource.TestCheckResourceAttr(resName, "cluster_size", "3"),
					resource.TestCheckResourceAttr(resName, "ssl_connection", "true"),
					resource.TestCheckResourceAttrSet(resName, "created"),
					resource.TestCheckResourceAttr(resName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resName, "engine", "postgresql"),
					resource.TestCheckResourceAttrSet(resName, "members.%"),
					resource.TestCheckResourceAttr(resName, "platform", "rdbms-default"),
					resource.TestCheckResourceAttrSet(resName, "port"),
					resource.TestCheckNoResourceAttr(resName, "oldest_restore_time"),
					// resource.TestCheckResourceAttr(resName, "status", "active"),
					resource.TestCheckResourceAttrSet(resName, "updated"),
					resource.TestCheckResourceAttrSet(resName, "version"),

					resource.TestCheckResourceAttr(resName, "allow_list.#", "1"),
					resource.TestCheckResourceAttr(resName, "allow_list.0", "10.0.0.4/32"),

					resource.TestCheckResourceAttrSet(resName, "hosts.primary"),
					resource.TestCheckResourceAttrSet(resName, "hosts.secondary"),

					resource.TestCheckResourceAttr(resName, "fork_source", "0"),
					resource.TestCheckResourceAttrSet(resName, "fork_restore_time"),

					resource.TestCheckResourceAttr(resName, "updates.hour_of_day", "2"),
					resource.TestCheckResourceAttr(resName, "updates.day_of_week", "3"),
					resource.TestCheckResourceAttr(resName, "updates.duration", "4"),
					resource.TestCheckResourceAttr(resName, "updates.frequency", "weekly"),

					resource.TestCheckResourceAttr(resName, "pending_updates.#", "0"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
