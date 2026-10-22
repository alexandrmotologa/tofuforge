package cost

import (
	"math"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

var ec2Pricing = map[string]float64{
	"t3.nano":     3.80,
	"t3.micro":    7.60,
	"t3.small":    15.20,
	"t3.medium":   30.40,
	"t3.large":    60.80,
	"t3.xlarge":   121.60,
	"t3.2xlarge":  243.20,
	"m5.large":    70.08,
	"m5.xlarge":   140.16,
	"m5.2xlarge":  280.32,
	"m6g.medium":  28.10,
	"m6g.large":   56.20,
	"m6g.xlarge":  112.40,
	"m6g.2xlarge": 224.80,
	"c5.large":    62.05,
	"c5.xlarge":   124.10,
}

var rdsPricing = map[string]float64{
	"db.t3.micro":    12.41,
	"db.t3.small":    24.82,
	"db.t3.medium":   49.64,
	"db.t3.large":    99.28,
	"db.m5.large":    124.10,
	"db.m5.xlarge":   248.20,
	"db.m6g.large":   109.50,
	"db.m6g.xlarge":  219.00,
	"db.m6g.2xlarge": 438.00,
	"db.r5.large":    175.20,
	"db.r5.xlarge":   350.40,
}

var gcpSQLPricing = map[string]float64{
	"db-f1-micro":       7.67,
	"db-g1-small":       25.55,
	"db-custom-2-7680":  74.50,
	"db-custom-4-16384": 149.00,
	"db-custom-8-32768": 298.00,
}

var gcpComputePricing = map[string]float64{
	"e2-micro":    6.11,
	"e2-small":    12.23,
	"e2-medium":   24.46,
	"e2-standard-2": 48.92,
	"e2-standard-4": 97.84,
}

var azureVMPricing = map[string]float64{
	"Standard_B1s":    7.59,
	"Standard_B2s":    30.37,
	"Standard_D2s_v5": 70.08,
	"Standard_D4s_v5": 140.16,
}

// EstimateNodeCost calculates the monthly cost before and after for an evaluated resource.
func EstimateNodeCost(node *model.ResourceNode) *model.CostEstimate {
	rType := strings.ToLower(node.Type)

	var beforeCost, afterCost float64
	var desc string

	switch {
	// AWS RDS
	case strings.Contains(rType, "aws_db_instance"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateRDSCost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateRDSCost(node.RawAfter)
		}
		desc = "AWS RDS instance & storage"

	// AWS EC2
	case strings.Contains(rType, "aws_instance"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateEC2Cost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateEC2Cost(node.RawAfter)
		}
		desc = "AWS EC2 compute instance"

	// AWS NAT Gateway
	case strings.Contains(rType, "aws_nat_gateway"):
		if node.Action != model.ActionCreate {
			beforeCost = 32.85
		}
		if node.Action != model.ActionDelete {
			afterCost = 32.85
		}
		desc = "AWS NAT Gateway ($32.85/mo flat)"

	// AWS ALB
	case strings.Contains(rType, "aws_lb") || strings.Contains(rType, "aws_alb"):
		if node.Action != model.ActionCreate {
			beforeCost = 16.20
		}
		if node.Action != model.ActionDelete {
			afterCost = 16.20
		}
		desc = "AWS Application Load Balancer"

	// AWS EBS Volume
	case strings.Contains(rType, "aws_ebs_volume"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateEBSCost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateEBSCost(node.RawAfter)
		}
		desc = "AWS EBS gp3 volume storage"

	// GCP Cloud SQL
	case strings.Contains(rType, "google_sql_database_instance"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateGCPSQLCost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateGCPSQLCost(node.RawAfter)
		}
		desc = "GCP Cloud SQL database tier"

	// GCP Compute
	case strings.Contains(rType, "google_compute_instance"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateGCPComputeCost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateGCPComputeCost(node.RawAfter)
		}
		desc = "GCP Compute Engine instance"

	// Azure VM
	case strings.Contains(rType, "azurerm_linux_virtual_machine") || strings.Contains(rType, "azurerm_windows_virtual_machine"):
		if node.Action != model.ActionCreate && node.RawBefore != nil {
			beforeCost = estimateAzureVMCost(node.RawBefore)
		}
		if node.Action != model.ActionDelete && node.RawAfter != nil {
			afterCost = estimateAzureVMCost(node.RawAfter)
		}
		desc = "Azure Virtual Machine size"

	default:
		return nil
	}

	delta := round(afterCost - beforeCost, 2)
	return &model.CostEstimate{
		MonthlyBefore: round(beforeCost, 2),
		MonthlyAfter:  round(afterCost, 2),
		MonthlyDelta:  delta,
		Currency:      "USD",
		Description:   desc,
	}
}

func estimateRDSCost(m map[string]interface{}) float64 {
	instanceClass, _ := m["instance_class"].(string)
	base := rdsPricing[instanceClass]
	if base == 0 {
		base = 45.00 // default fallback
	}

	storageGB := getFloat(m["allocated_storage"])
	if storageGB == 0 {
		storageGB = 20
	}
	storageCost := storageGB * 0.115

	return base + storageCost
}

func estimateEC2Cost(m map[string]interface{}) float64 {
	instanceType, _ := m["instance_type"].(string)
	price := ec2Pricing[instanceType]
	if price == 0 {
		price = 30.00
	}
	return price
}

func estimateEBSCost(m map[string]interface{}) float64 {
	size := getFloat(m["size"])
	if size == 0 {
		size = 20
	}
	return size * 0.08
}

func estimateGCPSQLCost(m map[string]interface{}) float64 {
	settings, ok := m["settings"].([]interface{})
	if ok && len(settings) > 0 {
		if smap, ok := settings[0].(map[string]interface{}); ok {
			tier, _ := smap["tier"].(string)
			if price, exists := gcpSQLPricing[tier]; exists {
				return price
			}
		}
	}
	return 50.00
}

func estimateGCPComputeCost(m map[string]interface{}) float64 {
	machineType, _ := m["machine_type"].(string)
	if price, exists := gcpComputePricing[machineType]; exists {
		return price
	}
	return 25.00
}

func estimateAzureVMCost(m map[string]interface{}) float64 {
	size, _ := m["size"].(string)
	if price, exists := azureVMPricing[size]; exists {
		return price
	}
	return 35.00
}

func getFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// ComputePlanCostAggregates sums before, after, and delta costs across all nodes.
func ComputePlanCostAggregates(nodes []model.ResourceNode) (beforeTotal, afterTotal, deltaTotal float64) {
	for _, n := range nodes {
		if n.Cost != nil {
			beforeTotal += n.Cost.MonthlyBefore
			afterTotal += n.Cost.MonthlyAfter
			deltaTotal += n.Cost.MonthlyDelta
		}
	}
	return round(beforeTotal, 2), round(afterTotal, 2), round(deltaTotal, 2)
}
