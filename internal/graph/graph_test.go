package graph

import (
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestBuildGraphAndBlastRadius(t *testing.T) {
	plan := &model.RawPlan{
		FormatVersion: "1.2",
		ResourceChanges: []model.RawResourceChange{
			{
				Address:      "aws_vpc.main",
				Type:         "aws_vpc",
				Name:         "main",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Change: model.RawChange{
					Actions: []string{"no-op"},
				},
			},
			{
				Address:      "aws_subnet.private",
				Type:         "aws_subnet",
				Name:         "private",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Change: model.RawChange{
					Actions: []string{"update"},
				},
			},
			{
				Address:      "aws_db_instance.primary",
				Type:         "aws_db_instance",
				Name:         "primary",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Change: model.RawChange{
					Actions: []string{"delete", "create"},
					ReplacePaths: [][]interface{}{
						{"allocated_storage"},
					},
				},
			},
			{
				Address:      "aws_ecs_service.api",
				Type:         "aws_ecs_service",
				Name:         "api",
				ProviderName: "registry.terraform.io/hashicorp/aws",
				Change: model.RawChange{
					Actions: []string{"update"},
				},
			},
		},
		Configuration: &model.RawConfiguration{
			RootModule: &model.ConfigModule{
				Resources: []model.ConfigResource{
					{
						Address: "aws_subnet.private",
						DependsOn: []string{"aws_vpc.main"},
					},
					{
						Address: "aws_db_instance.primary",
						DependsOn: []string{"aws_subnet.private"},
					},
					{
						Address: "aws_ecs_service.api",
						DependsOn: []string{"aws_db_instance.primary"},
					},
				},
			},
		},
	}

	g := BuildGraph(plan)

	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}

	if len(g.Edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(g.Edges))
	}

	alerts := g.ComputeBlastRadius()

	// aws_db_instance.primary is replace
	if len(alerts) != 1 {
		t.Fatalf("expected 1 destruction alert, got %d", len(alerts))
	}

	dbAlert := alerts[0]
	if dbAlert.ResourceAddress != "aws_db_instance.primary" {
		t.Errorf("expected alert for aws_db_instance.primary, got %s", dbAlert.ResourceAddress)
	}

	if dbAlert.Severity != model.SeverityCritical {
		t.Errorf("expected CRITICAL severity for stateful DB replacement, got %s", dbAlert.Severity)
	}

	if dbAlert.DownstreamCount != 1 {
		t.Errorf("expected 1 downstream resource (ecs_service), got %d", dbAlert.DownstreamCount)
	}

	if len(dbAlert.DownstreamPaths) != 1 || dbAlert.DownstreamPaths[0] != "aws_ecs_service.api" {
		t.Errorf("expected downstream path to contain aws_ecs_service.api, got %v", dbAlert.DownstreamPaths)
	}

	order, cycleFree := g.TopologicalSort()
	if !cycleFree {
		t.Errorf("expected cycle-free topological sort")
	}
	if len(order) != 4 {
		t.Errorf("expected 4 items in topological sort, got %d", len(order))
	}

	// aws_vpc.main should appear before aws_ecs_service.api
	vpcIdx := -1
	ecsIdx := -1
	for idx, addr := range order {
		if addr == "aws_vpc.main" {
			vpcIdx = idx
		}
		if addr == "aws_ecs_service.api" {
			ecsIdx = idx
		}
	}
	if vpcIdx >= ecsIdx {
		t.Errorf("expected vpc to precede ecs in topological sort, got vpcIdx=%d ecsIdx=%d", vpcIdx, ecsIdx)
	}
}
