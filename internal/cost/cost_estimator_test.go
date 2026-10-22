package cost

import (
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestEstimateNodeCost_RDSUpgrade(t *testing.T) {
	node := &model.ResourceNode{
		Address: "aws_db_instance.main",
		Type:    "aws_db_instance",
		Action:  model.ActionReplace,
		RawBefore: map[string]interface{}{
			"instance_class":    "db.m5.large", // $124.10
			"allocated_storage": 100,          // 100 * 0.115 = $11.50 -> $135.60
		},
		RawAfter: map[string]interface{}{
			"instance_class":    "db.m6g.xlarge", // $219.00
			"allocated_storage": 250,            // 250 * 0.115 = $28.75 -> $247.75
		},
	}

	costEst := EstimateNodeCost(node)
	if costEst == nil {
		t.Fatalf("expected cost estimate for RDS instance")
	}

	if costEst.MonthlyBefore != 135.60 {
		t.Errorf("expected MonthlyBefore 135.60, got %.2f", costEst.MonthlyBefore)
	}

	if costEst.MonthlyAfter != 247.75 {
		t.Errorf("expected MonthlyAfter 247.75, got %.2f", costEst.MonthlyAfter)
	}

	expectedDelta := 112.15
	if costEst.MonthlyDelta != expectedDelta {
		t.Errorf("expected MonthlyDelta %.2f, got %.2f", expectedDelta, costEst.MonthlyDelta)
	}
}

func TestEstimateNodeCost_CreateEC2(t *testing.T) {
	node := &model.ResourceNode{
		Address: "aws_instance.app",
		Type:    "aws_instance",
		Action:  model.ActionCreate,
		RawAfter: map[string]interface{}{
			"instance_type": "t3.medium",
		},
	}

	costEst := EstimateNodeCost(node)
	if costEst == nil {
		t.Fatalf("expected cost estimate for new EC2")
	}

	if costEst.MonthlyBefore != 0 {
		t.Errorf("expected 0 before on create, got %.2f", costEst.MonthlyBefore)
	}

	if costEst.MonthlyAfter != 30.40 {
		t.Errorf("expected 30.40 after, got %.2f", costEst.MonthlyAfter)
	}

	if costEst.MonthlyDelta != 30.40 {
		t.Errorf("expected +30.40 delta, got %.2f", costEst.MonthlyDelta)
	}
}
