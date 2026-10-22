package graph

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
)

// Graph represents the directed dependency graph of planned resources.
type Graph struct {
	Nodes    map[string]*model.ResourceNode
	Edges    []model.DependencyEdge
	InEdges  map[string][]string // target -> list of sources (upstream dependencies)
	OutEdges map[string][]string // source -> list of targets (downstream dependents)
}

// NewGraph initializes an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes:    make(map[string]*model.ResourceNode),
		Edges:    make([]model.DependencyEdge, 0),
		InEdges:  make(map[string][]string),
		OutEdges: make(map[string][]string),
	}
}

// AddNode adds a resource node to the graph.
func (g *Graph) AddNode(node model.ResourceNode) {
	g.Nodes[node.Address] = &node
}

// AddEdge adds a directed edge from source (dependency) to target (dependent).
func (g *Graph) AddEdge(source, target string, implicit bool, reason string) {
	if source == target || source == "" || target == "" {
		return
	}

	// Avoid duplicate edges
	for _, existing := range g.Edges {
		if existing.Source == source && existing.Target == target {
			return
		}
	}

	edgeID := fmt.Sprintf("%s->%s", source, target)
	edge := model.DependencyEdge{
		ID:       edgeID,
		Source:   source,
		Target:   target,
		Implicit: implicit,
		Reason:   reason,
	}

	g.Edges = append(g.Edges, edge)
	g.OutEdges[source] = append(g.OutEdges[source], target)
	g.InEdges[target] = append(g.InEdges[target], source)
}

// BuildGraph constructs the full dependency graph from raw plan data.
func BuildGraph(plan *model.RawPlan) *Graph {
	g := NewGraph()

	// 1. Create nodes for all resource changes
	for _, rc := range plan.ResourceChanges {
		action := parser.ClassifyAction(rc.Change.Actions)
		cat := parser.ClassifyCategory(rc.Type)
		stateful := parser.IsStatefulResource(rc.Type)
		diffs := parser.ComputeAttributeDiffs(rc.Change)

		node := model.ResourceNode{
			ID:             rc.Address,
			Address:        rc.Address,
			Name:           rc.Name,
			Type:           rc.Type,
			Provider:       parser.NormalizeProvider(rc.ProviderName),
			Module:         parser.ExtractModuleAddress(rc.Address),
			Mode:           rc.Mode,
			Action:         action,
			Category:       cat,
			IsStateful:     stateful,
			DirectImpact:   parser.IsDestructive(action),
			AttributeDiffs: diffs,
			RawBefore:      rc.Change.Before,
			RawAfter:       rc.Change.After,
		}

		g.AddNode(node)
	}

	// 2. Extract dependencies from Configuration if available
	if plan.Configuration != nil && plan.Configuration.RootModule != nil {
		extractConfigDependencies(g, plan.Configuration.RootModule, "")
	}

	// 3. Infer implicit dependencies between resources if configuration references are absent or sparse
	inferImplicitDependencies(g)

	// 4. Update Upstream and Downstream IDs on each node
	for addr, node := range g.Nodes {
		node.UpstreamIDs = g.InEdges[addr]
		node.Dependencies = g.InEdges[addr]
	}

	return g
}

// extractConfigDependencies traverses configuration module hierarchy to build dependency edges.
func extractConfigDependencies(g *Graph, mod *model.ConfigModule, modPrefix string) {
	if mod == nil {
		return
	}

	for _, res := range mod.Resources {
		targetAddr := res.Address
		if modPrefix != "" && !strings.HasPrefix(targetAddr, "module.") {
			targetAddr = modPrefix + "." + targetAddr
		}

		// Check explicit depends_on
		for _, dep := range res.DependsOn {
			sourceAddr := dep
			if modPrefix != "" && !strings.HasPrefix(sourceAddr, "module.") {
				sourceAddr = modPrefix + "." + sourceAddr
			}
			g.AddEdge(sourceAddr, targetAddr, false, "explicit depends_on")
		}

		// Check expressions references
		for _, exprVal := range res.Expressions {
			extractReferencesFromExpr(g, exprVal, targetAddr, modPrefix)
		}
	}

	// Recursively process child modules
	for callName, call := range mod.ModuleCalls {
		childPrefix := callName
		if modPrefix != "" {
			childPrefix = modPrefix + ".module." + callName
		} else {
			childPrefix = "module." + callName
		}

		if call.Module != nil {
			extractConfigDependencies(g, call.Module, childPrefix)
		}
	}
}

func extractReferencesFromExpr(g *Graph, expr interface{}, targetAddr, modPrefix string) {
	exprMap, ok := expr.(map[string]interface{})
	if !ok {
		return
	}

	if refs, hasRefs := exprMap["references"].([]interface{}); hasRefs {
		for _, ref := range refs {
			refStr, isStr := ref.(string)
			if !isStr {
				continue
			}

			// Clean reference: e.g. "aws_vpc.main.id" -> "aws_vpc.main"
			refResource := cleanReference(refStr)
			if refResource != "" {
				sourceAddr := refResource
				if modPrefix != "" && !strings.HasPrefix(sourceAddr, "module.") {
					sourceAddr = modPrefix + "." + sourceAddr
				}

				// Only add edge if the source node exists in the plan or matches known resource
				if _, exists := g.Nodes[sourceAddr]; exists {
					g.AddEdge(sourceAddr, targetAddr, false, "attribute reference")
				}
			}
		}
	}
}

// cleanReference strips trailing attribute fields like .id, .arn, .name from reference strings.
func cleanReference(ref string) string {
	parts := strings.Split(ref, ".")
	if len(parts) >= 2 {
		// e.g. aws_vpc.main.id -> aws_vpc.main
		// module.vpc.aws_subnet.private.id -> module.vpc.aws_subnet.private
		if parts[0] == "var" || parts[0] == "local" || parts[0] == "each" || parts[0] == "count" {
			return ""
		}
		if parts[0] == "module" {
			if len(parts) >= 4 {
				return strings.Join(parts[:4], ".")
			}
			return strings.Join(parts[:2], ".")
		}
		return fmt.Sprintf("%s.%s", parts[0], parts[1])
	}
	return ""
}

// inferImplicitDependencies creates edges based on well-known cloud resource hierarchy patterns.
func inferImplicitDependencies(g *Graph) {
	for targetAddr, targetNode := range g.Nodes {
		targetType := targetNode.Type

		for sourceAddr, sourceNode := range g.Nodes {
			if targetAddr == sourceAddr {
				continue
			}

			sourceType := sourceNode.Type

			// VPC -> Subnet
			if (sourceType == "aws_vpc" || sourceType == "google_compute_network" || sourceType == "azurerm_virtual_network") &&
				(strings.Contains(targetType, "subnet") || strings.Contains(targetType, "route_table") || strings.Contains(targetType, "security_group")) {
				if matchResourceAttributes(sourceNode, targetNode, "vpc_id", "network", "virtual_network_name") {
					g.AddEdge(sourceAddr, targetAddr, true, "network containment")
				}
			}

			// Security Group -> Security Group Rule
			if (sourceType == "aws_security_group" || sourceType == "azurerm_network_security_group") &&
				(strings.Contains(targetType, "security_group_rule") || strings.Contains(targetType, "network_security_rule")) {
				g.AddEdge(sourceAddr, targetAddr, true, "security group rule parent")
			}

			// Subnet / Security Group -> DB Instance or EC2 Instance
			if (strings.Contains(sourceType, "subnet") || strings.Contains(sourceType, "security_group")) &&
				(strings.Contains(targetType, "db_instance") || strings.Contains(targetType, "instance") || strings.Contains(targetType, "cluster")) {
				if matchResourceAttributes(sourceNode, targetNode, "subnet_ids", "vpc_security_group_ids", "security_groups") {
					g.AddEdge(sourceAddr, targetAddr, true, "compute/db network binding")
				}
			}

			// KMS Key -> S3 Bucket or RDS or EBS
			if strings.Contains(sourceType, "kms_key") &&
				(strings.Contains(targetType, "s3_bucket") || strings.Contains(targetType, "db_instance") || strings.Contains(targetType, "ebs_volume")) {
				if matchResourceAttributes(sourceNode, targetNode, "kms_key_id", "kms_master_key_id") {
					g.AddEdge(sourceAddr, targetAddr, true, "encryption key binding")
				}
			}
		}
	}
}

// matchResourceAttributes checks if any attribute in target before/after references source.
func matchResourceAttributes(source, target *model.ResourceNode, candidateKeys ...string) bool {
	// Simple heuristic check if source address/name appears in target values
	for _, m := range []map[string]interface{}{target.RawBefore, target.RawAfter} {
		if m == nil {
			continue
		}
		for _, key := range candidateKeys {
			if val, exists := m[key]; exists {
				valStr := fmt.Sprintf("%v", val)
				if strings.Contains(valStr, source.Name) || strings.Contains(valStr, source.Address) {
					return true
				}
			}
		}
	}
	return false
}
