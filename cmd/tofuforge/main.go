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

	"github.com/alexandrmotologa/tofuforge/internal/engine"
	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
	"github.com/alexandrmotologa/tofuforge/internal/reporter"
	"github.com/alexandrmotologa/tofuforge/internal/server"
)

var (
	version = "1.0.0"

	// inspect flags
	flagJSON           bool
	flagOutput         string
	flagNoColor        bool
	flagFailDestruct   bool
	flagFailCritical   bool
	flagFailSecurity   bool
	flagMaxBlastRadius int

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

	// 2. pr-comment command
	prCommentCmd := &cobra.Command{
		Use:   "pr-comment [plan.json]",
		Short: "Generate a compact GitHub/GitLab PR comment Markdown summary",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPRComment,
	}
	prCommentCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Write markdown to specified file instead of stdout")

	// 3. serve command
	serveCmd := &cobra.Command{
		Use:   "serve [plan.json]",
		Short: "Launch interactive web visualizer studio",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runServe,
	}
	serveCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Port for web visualizer studio")
	serveCmd.Flags().StringVar(&flagHost, "host", "localhost", "Host address to bind")
	serveCmd.Flags().BoolVar(&flagOpen, "open", false, "Automatically open visualizer in default browser")

	rootCmd.AddCommand(inspectCmd, prCommentCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadPlanFromArgOrStdin(args []string) (*model.AnalyzedPlan, error) {
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

	return engine.Analyze(raw), nil
}

func runInspect(cmd *cobra.Command, args []string) error {
	analyzed, err := loadPlanFromArgOrStdin(args)
	if err != nil {
		return err
	}

	// Output JSON if requested
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(analyzed)
	}

	// Output terminal report
	reporter.RenderTerminal(os.Stdout, analyzed, flagNoColor)

	// Write markdown output if requested
	if flagOutput != "" {
		md := reporter.GenerateMarkdownReport(analyzed)
		if err := os.WriteFile(flagOutput, []byte(md), 0644); err != nil {
			return fmt.Errorf("failed to write markdown output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Markdown PR summary written to: %s\n", flagOutput)
	}

	// CI Gating evaluation
	if flagFailDestruct && (analyzed.Stats.DeleteCount > 0 || analyzed.Stats.ReplaceCount > 0) {
		fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Destructive changes detected (--fail-on-destructive).")
		os.Exit(1)
	}

	if flagFailCritical && analyzed.Stats.CriticalAlertsCount > 0 {
		fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Critical destruction alerts detected (--fail-on-critical).")
		os.Exit(1)
	}

	if flagFailSecurity {
		for _, f := range analyzed.SecurityFindings {
			if f.Severity == model.SeverityCritical {
				fmt.Fprintln(os.Stderr, "\n[CI GATE FAILED] Critical security violations detected (--fail-on-security).")
				os.Exit(1)
			}
		}
	}

	if flagMaxBlastRadius > 0 && analyzed.Stats.TotalBlastRadius > flagMaxBlastRadius {
		fmt.Fprintf(os.Stderr, "\n[CI GATE FAILED] Total blast radius %d exceeds threshold %d (--max-blast-radius).\n",
			analyzed.Stats.TotalBlastRadius, flagMaxBlastRadius)
		os.Exit(1)
	}

	return nil
}

func runPRComment(cmd *cobra.Command, args []string) error {
	analyzed, err := loadPlanFromArgOrStdin(args)
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

func runServe(cmd *cobra.Command, args []string) error {
	var initialPlan *model.AnalyzedPlan
	if len(args) > 0 && args[0] != "" {
		p, err := loadPlanFromArgOrStdin(args)
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
