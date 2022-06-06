package ce_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestAccCEAnomalyMonitor_dimensionalserial limits the number of parallel tests run with a type of DIMENSIONAL to 1.
// This is required as AWS only allows 1 Anomaly Monitor with a type of DIMENSIONAL per AWS account.
func TestAccCEAnomalyMonitor_dimensionalserial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ContainerService": {
			"basic":      testAccCEAnomalyMonitor_basic,
			"disappears": testAccCEAnomalyMonitor_disappears,
			"name":       testAccCEAnomalyMonitor_Name,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccCEAnomalyMonitor_basic(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dimensionValue := "SERVICE"
	dimensionBadValue := "BADVALUE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCEAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config:      testAccCEAnomalyMonitorConfig(rName, dimensionBadValue),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`expected dimension to be one of \[SERVICE\], got %s`, dimensionBadValue)),
			},
			{
				Config: testAccCEAnomalyMonitorConfig(rName, dimensionValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCEAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCEAnomalyMonitor_Name(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	dimensionValue := "SERVICE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCEAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCEAnomalyMonitorConfig(rName, dimensionValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCEAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCEAnomalyMonitorConfig(rName2, dimensionValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCEAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccCEAnomalyMonitor_Custom(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCEAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCEAnomalyMonitorConfig_Custom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCEAnomalyMonitorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCEAnomalyMonitor_disappears(t *testing.T) {
	resourceName := "aws_ce_anomaly_monitor.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCEAnomalyMonitorDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCEAnomalyMonitorConfig(rName, "SERVICE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCEAnomalyMonitorExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfce.ResourceAnomalyMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCEAnomalyMonitorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Database ID is set")
		}

		resp, err := conn.GetAnomalyMonitors(&costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return err
		}

		if resp == nil || len(resp.AnomalyMonitors) < 1 {
			return fmt.Errorf("Anomaly Monitor (%s) not found", rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func testAccCheckCEAnomalyMonitorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_monitor" {
			continue
		}

		resp, err := conn.GetAnomalyMonitors(&costexplorer.GetAnomalyMonitorsInput{MonitorArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalyMonitor, rs.Primary.ID, err)
		}

		if resp != nil && len(resp.AnomalyMonitors) > 0 {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalyMonitor, rs.Primary.ID, errors.New("still exists"))
		}
	}

	return nil

}

func testAccCEAnomalyMonitorConfig(rName string, dimension string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name      = %[1]q
  type      = "DIMENSIONAL"
  dimension = %[2]q
}
`, rName, dimension)
}

func testAccCEAnomalyMonitorConfig_Custom(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON
}
`, rName)
}
