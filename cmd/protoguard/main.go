package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codec404/protoguard/pkg/protoguard"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRoot().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "protoguard",
		Short:         "Diff OpenAPI / protobuf contracts with structured output and optional LLM explanations",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newDiffCmd())
	return root
}

func newDiffCmd() *cobra.Command {
	var specStr string
	var maxSpecMB int

	var d struct {
		protoguard.Options
		Format         string
		PrettyJSON     bool
		FailOnBreaking bool
	}

	d.Spec = protoguard.SpecAuto
	d.Format = "both"
	d.FailOnBreaking = true

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two OpenAPI or protobuf FileDescriptorSet artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if d.OldPath == "" || d.NewPath == "" {
				return fmt.Errorf("require --old and --new")
			}
			if maxSpecMB > 512 {
				return fmt.Errorf("--max-spec-mb cannot exceed 512")
			}
			if maxSpecMB < 1 {
				maxSpecMB = 32
			}
			d.MaxSpecBytes = int64(maxSpecMB) * 1024 * 1024
			d.Spec = protoguard.SpecKind(specStr)

			res, err := protoguard.DiffWithStderr(cmd.Context(), d.Options, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			switch d.Format {
			case "json":
				if err := protoguard.WriteReportJSON(cmd.OutOrStdout(), res, d.PrettyJSON); err != nil {
					return err
				}
			case "markdown", "md":
				protoguard.WriteReportMarkdown(cmd.OutOrStdout(), res)
			case "both":
				if err := protoguard.WriteReportJSON(cmd.OutOrStdout(), res, d.PrettyJSON); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout())
				protoguard.WriteReportMarkdown(cmd.OutOrStdout(), res)
			default:
				return fmt.Errorf("invalid --format %q (json|markdown|both)", d.Format)
			}

			if d.FailOnBreaking && res.Breaking {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&d.OldPath, "old", "", "Path to old spec (OpenAPI YAML/JSON or .pb FileDescriptorSet)")
	cmd.Flags().StringVar(&d.NewPath, "new", "", "Path to new spec")
	cmd.Flags().StringVar(&specStr, "spec", "auto", "openapi | protobuf | auto")
	cmd.Flags().IntVar(&maxSpecMB, "max-spec-mb", 32, "Maximum size in MiB per spec file (bounds memory use)")
	cmd.Flags().BoolVar(&d.AllowOpenAPIExternalRefs, "allow-openapi-external-refs", false, "Resolve remote OpenAPI $ref URLs (SSRF risk — off by default)")
	cmd.Flags().StringVar(&d.Format, "format", "both", "json | markdown | both")
	cmd.Flags().BoolVar(&d.PrettyJSON, "pretty-json", false, "Indent JSON output")
	cmd.Flags().BoolVar(&d.SkipLLM, "skip-llm", false, "Structured diff only (no HTTP calls)")
	cmd.Flags().BoolVar(&d.IncludeFullSpec, "include-full-spec", false, "Debug marker only; prompts stay structured-diff-only")
	cmd.Flags().BoolVar(&d.RedactURLs, "redact-urls", false, "Strip OpenAPI servers before diff")
	cmd.Flags().BoolVar(&d.FailOnBreaking, "fail-on-breaking", true, "Exit 1 when report contains BREAKING changes")
	cmd.Flags().StringVar(&d.CacheDir, "cache-dir", "", "LLM response cache directory (default XDG_CACHE_HOME/protoguard)")

	cmd.Flags().StringVar(&d.LLM.Target, "llm-target", "", "local | cloud (default local; env PROTOGUARD_LLM_TARGET)")
	cmd.Flags().StringVar(&d.LLM.BaseURL, "llm-base-url", "", "Chat API base URL (http/https only; no user:pass in URL)")
	cmd.Flags().StringVar(&d.LLM.Model, "llm-model", "", "Model name (local default llama3.2 if unset)")
	cmd.Flags().StringVar(&d.LLM.APIKey, "llm-api-key", "", "Bearer token when required (prefer PROTOGUARD_LLM_API_KEY)")
	cmd.Flags().StringVar(&d.LLM.Backend, "llm-backend", "openai_shape", "openai_shape | ollama")
	cmd.Flags().IntVar(&d.LLM.MaxOutputTokens, "llm-max-output-tokens", 2048, "Max completion tokens per chunk")
	cmd.Flags().IntVar(&d.LLM.TimeoutSecs, "llm-timeout-secs", 120, "HTTP timeout per chunk")
	cmd.Flags().IntVar(&d.LLM.MaxChunkBytes, "llm-max-chunk-bytes", 8000, "Approximate max chunk payload size")

	return cmd
}
