package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kube-agents-test run [flags] <scenario-file-or-dir>")
	}
	if args[0] != "run" {
		return fmt.Errorf("unknown command %q (only 'run' is supported)", args[0])
	}
	return runCommand(args[1:])
}

func runCommand(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	kubeconfig := fs.String("kubeconfig", "", "path to kubeconfig for attached mode")
	clusterMode := fs.String("cluster-mode", "ephemeral", "cluster mode: ephemeral or attached")
	kindClusterName := fs.String("kind-cluster-name", "", "name for ephemeral kind cluster")
	agentMode := fs.String("agent-mode", "pods", "agent deploy mode: pods or local")
	agentConfigPath := fs.String("agent-config", "", "path to agent registry config (format TBD)")
	timeoutOverride := fs.Duration("timeout", 0, "override scenario expect timeout (0 keeps YAML value)")
	verbose := fs.Bool("v", false, "verbose logging")

	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return fmt.Errorf("run requires exactly one path argument")
	}
	path := rest[0]

	mode, err := cluster.ParseMode(*clusterMode)
	if err != nil {
		return err
	}
	deployMode, err := agent.ParseDeployMode(*agentMode)
	if err != nil {
		return err
	}
	_ = *agentConfigPath // reserved until agent config file format is defined
	_ = *verbose

	loaded, err := runner.LoadPath(path)
	if err != nil {
		return err
	}
	if *timeoutOverride > 0 {
		for i := range loaded {
			loaded[i].Scenario.Expect.Timeout = *timeoutOverride
		}
	}

	r, err := runner.New(runner.Config{
		ClusterMode: mode,
		Kubeconfig:  *kubeconfig,
		EphemeralOpts: cluster.EphemeralOptions{
			Backend:     "kind",
			ClusterName: *kindClusterName,
		},
		AgentConfig: agent.Config{Mode: deployMode},
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	suite, err := r.Run(ctx, loaded)
	if err != nil {
		return err
	}

	printSummary(suite)
	if suite.Failed > 0 {
		return errors.New("one or more scenarios failed")
	}
	return nil
}

func printSummary(suite *runner.SuiteResult) {
	for _, res := range suite.Results {
		status := "PASS"
		if !res.Passed || res.Err != nil {
			status = "FAIL"
		}
		line := fmt.Sprintf("%s %s", status, res.Name)
		if res.Err != nil {
			line += fmt.Sprintf(" (%v)", res.Err)
		}
		fmt.Println(line)
	}
	fmt.Printf("\n%d passed, %d failed\n", suite.Passed, suite.Failed)
	if suite.Failed > 0 {
		fmt.Fprintln(os.Stderr, strings.TrimSpace("scenario failures detected"))
	}
}
