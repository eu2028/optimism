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
    using stdJson for string;

    string internal _json;

    /// @notice Deploy all of the L1 contracts necessary for a full Superchain with a single Op Chain.

    function run() public override {
        console.log("Deploying fresh Superchain Shared contracts");

        // ----- Deploy the Superchain Shared contracts -----
        // TODO: replace this with a call to `op-deployer bootstrap superchain`
        // Note: for the moment deploySuperchain() is unnecessary, because testing is using a fork of mainnet,
        //       so
        deploySuperchain();

        // ----- Deploy the OPCM along with implementations contracts it requires -----
        // WIP: this currently only works when anvil is run by forking mainnet, otherwise it fails with:
        // "error getting superchain config: unsupported chain ID: 31337\n".
        // One option for fixing would be to add 31337 to a standard-versions toml file
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

        bytes memory result = Process.run(cmds);
        console.log(string(result));

        string[] memory cmds2 = new string[](7);
        cmds2[0] = "scripts/op-deployer";
        cmds2[1] = "apply";
        cmds2[2] = "--l1-rpc-url";
        cmds2[3] = "http://localhost:8545";
        cmds2[4] = "--private-key";
        // Private key for first default hardhat account
        cmds2[5] = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";

        bytes memory result2 = Process.run(cmds2);
        console.log(string(result2));
        // TODO: Save OPCM addresses from state.json

        console.log("DeployConfig: reading file %s", "state.json");
        try vm.readFile("state.json") returns (string memory data_) {
            _json = data_;
        } catch {
            require(false, string.concat("DeployConfig: cannot find deploy config file at ", "state.json"));
        }

        save("SuperchainConfig", _json.readAddress("$.opChainDeployments[0].superchainConfigImplAddress"));
        save("SuperchainConfigProxy", _json.readAddress("$.opChainDeployments[0].superchainConfigProxyAddress"));
        save("ProtocolVersions", _json.readAddress("$.opChainDeployments[0].protocolVersions"));
        save("ProtocolVersionsProxy", _json.readAddress("$.opChainDeployments[0].protocolVersionsProxy"));
        save("SuperchainProxyAdmin", _json.readAddress("$.opChainDeployments[0].superchainProxyAdmin"));
        save("SuperchainConfigProxy", _json.readAddress("$.opChainDeployments[0].superchainConfigProxy"));
        save("SuperchainConfig", _json.readAddress("$.opChainDeployments[0].superchainConfig"));
        save("ProtocolVersionsProxy", _json.readAddress("$.opChainDeployments[0].protocolVersionsProxy"));
        save("ProtocolVersions", _json.readAddress("$.opChainDeployments[0].protocolVersions"));
        save("L1CrossDomainMessenger", _json.readAddress("$.opChainDeployments[0].l1CrossDomainMessenger"));
        save("OptimismMintableERC20Factory", _json.readAddress("$.opChainDeployments[0].optimismMintableERC20Factory"));
        save("SystemConfig", _json.readAddress("$.opChainDeployments[0].systemConfig"));
        save("L1StandardBridge", _json.readAddress("$.opChainDeployments[0].l1StandardBridge"));
        save("L1ERC721Bridge", _json.readAddress("$.opChainDeployments[0].l1ERC721Bridge"));
        save("OptimismPortal2", _json.readAddress("$.opChainDeployments[0].optimismPortal2"));
        save("DisputeGameFactory", _json.readAddress("$.opChainDeployments[0].disputeGameFactory"));
        save("DelayedWETH", _json.readAddress("$.opChainDeployments[0].delayedWETH"));
        save("PreimageOracle", _json.readAddress("$.opChainDeployments[0].preimageOracle"));
        save("Mips", _json.readAddress("$.opChainDeployments[0].mips"));
        save("OPContractsManager", _json.readAddress("$.opChainDeployments[0].oPContractsManager"));
        save("ProxyAdmin", _json.readAddress("$.opChainDeployments[0].proxyAdmin"));
        save("AddressManager", _json.readAddress("$.opChainDeployments[0].addressManager"));
        save("L1ERC721BridgeProxy", _json.readAddress("$.opChainDeployments[0].l1ERC721BridgeProxy"));
        save("SystemConfigProxy", _json.readAddress("$.opChainDeployments[0].systemConfigProxy"));
        save(
            "OptimismMintableERC20FactoryProxy",
            _json.readAddress("$.opChainDeployments[0].optimismMintableERC20FactoryProxy")
        );
        save("L1StandardBridgeProxy", _json.readAddress("$.opChainDeployments[0].l1StandardBridgeProxy"));
        save("L1CrossDomainMessengerProxy", _json.readAddress("$.opChainDeployments[0].l1CrossDomainMessengerProxy"));
        save("DisputeGameFactoryProxy", _json.readAddress("$.opChainDeployments[0].disputeGameFactoryProxy"));
        save("PermissionedDelayedWETHProxy", _json.readAddress("$.opChainDeployments[0].permissionedDelayedWETHProxy"));
        save("AnchorStateRegistryProxy", _json.readAddress("$.opChainDeployments[0].anchorStateRegistryProxy"));
        save("AnchorStateRegistry", _json.readAddress("$.opChainDeployments[0].anchorStateRegistry"));
        save("PermissionedDisputeGame", _json.readAddress("$.opChainDeployments[0].permissionedDisputeGame"));
        save("OptimismPortalProxy", _json.readAddress("$.opChainDeployments[0].optimismPortalProxy"));
    }
}
