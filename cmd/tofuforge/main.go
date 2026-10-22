package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/alexandrmotologa/tofuforge/internal/config"
	"github.com/alexandrmotologa/tofuforge/internal/diff"
	"github.com/alexandrmotologa/tofuforge/internal/engine"
	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
	"github.com/alexandrmotologa/tofuforge/internal/reporter"
	"github.com/alexandrmotologa/tofuforge/internal/server"
)

var (
	version = "1.1.0"

	flagConfig string

	// inspect flags
	flagJSON           bool
	flagOutput         string
	flagNoColor        bool
	flagFailDestruct   bool
	flagFailCritical   bool
	flagFailSecurity   bool
	flagMaxBlastRadius int
	flagMaxCostDelta   float64

	// report flags
	flagReportOutput string

	// serve flags
	flagPort int
	flagHost string
	flagOpen bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "tofuforge",
		Short:   "TofuForge: Interactive Terraform & OpenTofu Plan Visualizer & Destruction Guard",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", "", "Path to .tofuforge.yaml policy configuration")

	// 1. inspect command
	inspectCmd := &cobra.Command{
		Use:   "inspect [plan.json]",
		Short: "Inspect a plan JSON file, audit security, calculate blast radius, and enforce CI gates",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInspect,
	}
	inspectCmd.Flags().BoolVar(&flagJSON, "json", false, "Output analyzed results as JSON")
	inspectCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write markdown PR summary to specified file")
	inspectCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colorized terminal output")
	inspectCmd.Flags().BoolVar(&flagFailDestruct, "fail-on-destructive", false, "Exit code 1 if any resource is deleted or replaced")
	inspectCmd.Flags().BoolVar(&flagFailCritical, "fail-on-critical", false, "Exit code 1 if any critical destruction alerts exist")
	inspectCmd.Flags().BoolVar(&flagFailSecurity, "fail-on-security", false, "Exit code 1 if any critical security violations exist")
	inspectCmd.Flags().IntVar(&flagMaxBlastRadius, "max-blast-radius", 0, "Exit code 1 if total blast radius score exceeds threshold (0 disables)")
	inspectCmd.Flags().Float64Var(&flagMaxCostDelta, "max-cost-delta", 0, "Exit code 1 if monthly cost delta exceeds threshold in USD (0 disables)")

	// 2. pr-comment command
	prCommentCmd := &cobra.Command{
		Use:   "pr-comment [plan.json]",
		Short: "Generate a compact GitHub/GitLab PR comment Markdown summary",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPRComment,
	}
	prCommentCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write markdown to specified file instead of stdout")

	// 3. diff command (Plan-to-Plan comparison)
	diffCmd := &cobra.Command{
		Use:   "diff <base-plan.json> <target-plan.json>",
		Short: "Compare two plan revisions to verify resolved destruction risks and delta changes",
		Args:  cobra.ExactArgs(2),
		RunE:  runDiff,
	}
	diffCmd.Flags().BoolVar(&flagJSON, "json", false, "Output comparison result as JSON")
	diffCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write markdown comparison to specified file")
	diffCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colorized terminal output")

	// 4. report command (Standalone Single-File HTML)
	reportCmd := &cobra.Command{
		Use:   "report [plan.json]",
		Short: "Generate a standalone single-file offline HTML visualizer report",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runReport,
	}
	reportCmd.Flags().StringVarP(&flagReportOutput, "output", "o", "tofuforge-report.html", "Path to save standalone HTML report")

	// 5. serve command
	serveCmd := &cobra.Command{
		Use:   "serve [plan.json]",
		Short: "Launch interactive web visualizer studio",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runServe,
	}
	serveCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Port for web visualizer studio")
	serveCmd.Flags().StringVar(&flagHost, "host", "localhost", "Host address to bind")
	serveCmd.Flags().BoolVar(&flagOpen, "open", false, "Automatically open visualizer in default browser")

	rootCmd.AddCommand(inspectCmd, prCommentCmd, diffCmd, reportCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadPlanFromArgOrStdin(args []string, cfg *config.Config) (*model.AnalyzedPlan, error) {
	var reader io.Reader

	if len(args) == 0 || args[0] == "-" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("no input plan provided: supply a file path or pipe JSON to stdin")
		}
		reader = os.Stdin
	} else {
		file, err := os.Open(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to open plan file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	raw, err := parser.ParsePlan(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	return engine.AnalyzeWithConfig(raw, cfg), nil
}

func runInspect(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		return err
	}

	analyzed, err := loadPlanFromArgOrStdin(args, cfg)
	if err != nil {
		return err
	}

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(analyzed)
	}

	reporter.RenderTerminal(os.Stdout, analyzed, flagNoColor)

	if flagOutput != "" {
		md := reporter.GenerateMarkdownReport(analyzed)
		if err := os.WriteFile(flagOutput, []byte(md), 0644); err != nil {
			return fmt.Errorf("failed to write markdown output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Markdown PR summary written to: %s\n", flagOutput)
	}

	// Apply configuration strictness overrides if flags not explicitly set
	failDestruct := flagFailDestruct || cfg.Strictness.FailOnDestructive
	failCritical := flagFailCritical || cfg.Strictness.FailOnCritical
	failSecurity := flagFailSecurity || cfg.Strictness.FailOnSecurity
	maxBlast := flagMaxBlastRadius
	if maxBlast == 0 && cfg.Strictness.MaxBlastRadius > 0 {
		maxBlast = cfg.Strictness.MaxBlastRadius
	}
	maxCost := flagMaxCostDelta
	if maxCost == 0 && cfg.Strictness.MaxCostDelta > 0 {
		maxCost = cfg.Strictness.MaxCostDelta
	}

	if failDestruct && (analyzed.Stats.DeleteCount > 0 || analyzed.Stats.ReplaceCount > 0) {
		fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Destructive changes detected (--fail-on-destructive).")
		os.Exit(1)
	}

	if failCritical && analyzed.Stats.CriticalAlertsCount > 0 {
		fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Critical destruction alerts detected (--fail-on-critical).")
		os.Exit(1)
	}

	if failSecurity {
		for _, f := range analyzed.SecurityFindings {
			if f.Severity == model.SeverityCritical {
				fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Critical security violations detected (--fail-on-security).")
				os.Exit(1)
			}
		}
	}

	if maxBlast > 0 && analyzed.Stats.TotalBlastRadius > maxBlast {
		fmt.Fprintf(os.Stderr, "\n[CI GATE FAILED] Total blast radius %d exceeds threshold %d (--max-blast-radius).\n",
			analyzed.Stats.TotalBlastRadius, maxBlast)
		os.Exit(1)
	}

	if maxCost > 0 && analyzed.Stats.TotalMonthlyDelta > maxCost {
		fmt.Fprintf(os.Stderr, "\n[CI GATE FAILED] Monthly cost increase +$%.2f exceeds limit $%.2f (--max-cost-delta).\n",
			analyzed.Stats.TotalMonthlyDelta, maxCost)
		os.Exit(1)
	}

	return nil
}

func runPRComment(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		return err
	}

	analyzed, err := loadPlanFromArgOrStdin(args, cfg)
	if err != nil {
		return err
	}

	md := reporter.GenerateMarkdownReport(analyzed)

	if flagOutput != "" {
		if err := os.WriteFile(flagOutput, []byte(md), 0644); err != nil {
			return fmt.Errorf("failed to write markdown file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Markdown written to %s\n", flagOutput)
		return nil
	}

	fmt.Print(md)
	return nil
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		return err
	}

	basePlan, err := loadPlanFromFile(args[0], cfg)
	if err != nil {
		return fmt.Errorf("failed to load base plan: %w", err)
	}

	targetPlan, err := loadPlanFromFile(args[1], cfg)
	if err != nil {
		return fmt.Errorf("failed to load target plan: %w", err)
	}

	comp := diff.ComparePlans(basePlan, targetPlan)

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(comp)
	}

	diff.RenderTerminalDiff(os.Stdout, comp, flagNoColor)

	if flagOutput != "" {
		md := diff.GenerateMarkdownDiff(comp)
		if err := os.WriteFile(flagOutput, []byte(md), 0644); err != nil {
			return fmt.Errorf("failed to write diff markdown: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Comparison markdown written to: %s\n", flagOutput)
	}

	return nil
}

func runReport(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		return err
	}

	analyzed, err := loadPlanFromArgOrStdin(args, cfg)
	if err != nil {
		return err
	}

	htmlContent, err := reporter.GenerateStandaloneHTML(analyzed)
	if err != nil {
		return fmt.Errorf("failed to generate standalone HTML: %w", err)
	}

	outputPath := flagReportOutput
	if outputPath == "" {
		outputPath = "tofuforge-report.html"
	}

	if err := os.WriteFile(outputPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}

	fmt.Printf("Standalone visual HTML report successfully generated at: %s\n", outputPath)
	return nil
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, _ := config.LoadConfig(flagConfig)

	var initialPlan *model.AnalyzedPlan
	if len(args) > 0 && args[0] != "" {
		p, err := loadPlanFromArgOrStdin(args, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not preload plan '%s': %v\n", args[0], err)
		} else {
			initialPlan = p
			fmt.Printf("Loaded initial plan: %d resources, %d alerts\n",
				initialPlan.Stats.TotalResources, len(initialPlan.DestructionAlerts))
		}
	}

	srv := server.NewServer(initialPlan)
	addr := fmt.Sprintf("%s:%d", flagHost, flagPort)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("\n🚀 TofuForge Web Studio running at: %s\n", url)
	fmt.Println("Press Ctrl+C to terminate.")

	if flagOpen {
		openBrowser(url)
	}

	return http.ListenAndServe(addr, srv.Handler())
}

func loadPlanFromFile(filePath string, cfg *config.Config) (*model.AnalyzedPlan, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	raw, err := parser.ParsePlan(file)
	if err != nil {
		return nil, err
	}

	return engine.AnalyzeWithConfig(raw, cfg), nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
