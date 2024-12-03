// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { VmSafe } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";

// Scripts
import { Deploy } from "scripts/deploy/Deploy.s.sol";
import { DeploySuperchainInput, DeploySuperchain, DeploySuperchainOutput } from "scripts/deploy/DeploySuperchain.s.sol";

// Libraries
import { Types } from "scripts/libraries/Types.sol";
import { Process } from "scripts/libraries/Process.sol";

/// @title Upgrade
contract Upgrade is Deploy {
    /// @notice Deploy all of the L1 contracts necessary for a full Superchain with a single Op Chain.
    function run() public override {
        console.log("Deploying fresh Superchain Shared contracts");

        // Deploy the Superchain Shared contracts
        // TODO: replace this with a call to `op-deployer bootstrap superchain`
        deploySuperchain();

        string[] memory cmds = new string[](9);
        cmds[0] = "scripts/op-deployer";
        cmds[1] = "bootstrap";
        cmds[2] = "opcm";
        cmds[3] = "--artifacts-locator";
        cmds[4] = "tag://op-contracts/v1.6.0";
        cmds[5] = "--l1-rpc-url";
        cmds[6] = "http://localhost:8545";
        cmds[7] = "--private-key";
        // Private key for first default hardhat account
        cmds[8] = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";

        // WIP: fails with error getting superchain config: unsupported chain ID: 31337\n")] setUp() (gas: 0) bytes
        // TODO: Hook into deploy superchain output and generate a toml file for op-deployer to read from?
        bytes memory result = Process.run(cmds);
        console.log(string(result));
    }
}
