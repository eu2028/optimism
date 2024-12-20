package report

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

var (
	DeploymentTxHashFlag = &cli.StringFlag{
		Name:    "deployment-tx-hash",
		Usage:   "The transaction hash of the deployment.",
		EnvVars: deployer.PrefixEnvVar("DEPLOYMENT_TX_HASH"),
	}
	ReleaseFlag = &cli.StringFlag{
		Name:    "release",
		Usage:   "The release tag of the deployment, of the format 'op-contracts/vX.Y.Z'.",
		EnvVars: deployer.PrefixEnvVar("RELEASE"),
	}
)

var L1ReportFlags = []cli.Flag{
	deployer.L1RPCURLFlag,
	deployer.OutfileFlag,
	DeploymentTxHashFlag,
	ReleaseFlag,
}

var Commands = cli.Commands{
	{
		Name:   "l1",
		Usage:  "Generates a report for a chain whose L1 contracts were deployed using op-deployer.",
		Flags:  L1ReportFlags,
		Action: ScanL1CLI,
	},
}

func ScanL1CLI(cliCtx *cli.Context) error {
	rpcURL := cliCtx.String(deployer.L1RPCURLFlag.Name)
	if rpcURL == "" {
		return fmt.Errorf("l1-rpc-url is required")
	}
	outfile := cliCtx.String(deployer.OutfileFlag.Name)
	if outfile == "" {
		return fmt.Errorf("outfile is required")
	}
	deploymentTxHashStr := cliCtx.String(DeploymentTxHashFlag.Name)
	if deploymentTxHashStr == "" {
		return fmt.Errorf("deployment-tx-hash is required")
	}
	deploymentTxHash := common.HexToHash(deploymentTxHashStr)
	release := cliCtx.String(ReleaseFlag.Name)
	if release == "" {
		return fmt.Errorf("release is required")
	}

	ctx, cancel := context.WithTimeout(cliCtx.Context, 5*time.Minute)
	defer cancel()

	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to dial RPC client: %w", err)
	}

	report, err := ScanL1(ctx, rpcClient, deploymentTxHash, release)
	if err != nil {
		return fmt.Errorf("failed to scan L1: %w", err)
	}

	if err := jsonutil.WriteJSON(report, ioutil.ToStdOutOrFileOrNoop(outfile, 0o755)); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}
