// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { VmSafe } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";

// Scripts
import { Deployer } from "scripts/deploy/Deployer.sol";
import { DeploySuperchainInput, DeploySuperchain, DeploySuperchainOutput } from "scripts/deploy/DeploySuperchain.s.sol";

// Libraries
import { Types } from "scripts/libraries/Types.sol";
import { Process } from "scripts/libraries/Process.sol";

/// @title Upgrade
contract Upgrade is Deployer {
    using stdJson for string;

    string internal _stateJson;
    string internal _stdVersionsToml;

    /// @notice Deploy all of the L1 contracts necessary for a full Superchain with a single Op Chain.

    function run() public {
        // Note: For now we are using a fork of mainnet, so we don't need to deploy the Superchain Shared contracts.
        //       This will change after the next release when we have all implementations in a single release.

        // ----- Deploy the OPCM along with implementations contracts it requires -----
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

        string[] memory cmds2 = new string[](8);
        cmds2[0] = "scripts/op-deployer";
        cmds2[1] = "apply";
        cmds2[2] = "--l1-rpc-url";
        cmds2[3] = "http://localhost:8545";
        cmds2[4] = "--private-key";
        // Private key for first default hardhat account
        cmds2[5] = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";
        cmds2[6] = "--workdir";
        cmds2[7] = "test/fixtures";

        bytes memory result2 = Process.run(cmds2);
        console.log(string(result2));

        console.log("Upgrade: reading file state.json");
        try vm.readFile("test/fixtures/state.json") returns (string memory data_) {
            _stateJson = data_;
        } catch {
            require(false, "Upgrade: cannot find state.json");
        }

        // Save superchain shared contracts
        save("SuperchainConfig", _stateJson.readAddress("$.superchainDeployment.superchainConfigImplAddress"));
        save("SuperchainConfigProxy", _stateJson.readAddress("$.superchainDeployment.superchainConfigProxyAddress"));
        save("ProtocolVersions", _stateJson.readAddress("$.superchainDeployment.protocolVersionsImplAddress"));
        save("ProtocolVersionsProxy", _stateJson.readAddress("$.superchainDeployment.protocolVersionsProxyAddress"));
        save("SuperchainProxyAdmin", _stateJson.readAddress("$.superchainDeployment.proxyAdminAddress"));

        console.log("Upgrade: reading file %s", "test/fixtures/standard-versions.toml");
        try vm.readFile("test/fixtures/standard-versions.toml") returns (string memory data_) {
            _stdVersionsToml = data_;
        } catch {
            require(false, "Upgrade: cannot find standard-versions.toml");
        }

        vm.parseToml(_stdVersionsToml, ".releases[\"op-contracts/v1.6.0\"]");
        // Save OPCM contracts
        save(
            "L1CrossDomainMessenger",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].l1_cross_domain_messenger.implementation_address")
            )
        );
        save(
            "OptimismMintableERC20Factory",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(
                    ".releases[\"op-contracts/v1.6.0\"].optimism_mintable_erc20_factory.implementation_address"
                )
            )
        );
        save(
            "SystemConfig",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].system_config.implementation_address")
            )
        );
        save(
            "L1StandardBridge",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].l1_standard_bridge.implementation_address")
            )
        );
        save(
            "L1ERC721Bridge",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].l1_erc721_bridge.implementation_address")
            )
        );
        save(
            "OptimismPortal2",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].optimism_portal.implementation_address")
            )
        );
        save(
            "DisputeGameFactory",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].dispute_game_factory.implementation_address")
            )
        );
        save(
            "DelayedWETHProxy",
            vm.parseTomlAddress(
                _stdVersionsToml,
                string.concat(".releases[\"op-contracts/v1.6.0\"].delayed_weth.implementation_address")
            )
        );
        save(
            "PreimageOracle",
            vm.parseTomlAddress(
                _stdVersionsToml, string.concat(".releases[\"op-contracts/v1.6.0\"].preimage_oracle.address")
            )
        );
        save(
            "Mips",
            vm.parseTomlAddress(_stdVersionsToml, string.concat(".releases[\"op-contracts/v1.6.0\"].mips.address"))
        );

        // save("OPContractsManager", _json.readAddress("$.opChainDeployments[0].opContractsManager"));

        save("ProxyAdmin", _stateJson.readAddress("$.opChainDeployments[0].proxyAdminAddress"));
        save("AddressManager", _stateJson.readAddress("$.opChainDeployments[0].addressManagerAddress"));
        save("L1ERC721BridgeProxy", _stateJson.readAddress("$.opChainDeployments[0].l1ERC721BridgeProxyAddress"));
        save("SystemConfigProxy", _stateJson.readAddress("$.opChainDeployments[0].systemConfigProxyAddress"));
        save(
            "OptimismMintableERC20FactoryProxy",
            _stateJson.readAddress("$.opChainDeployments[0].optimismMintableERC20FactoryProxyAddress")
        );
        save("L1StandardBridgeProxy", _stateJson.readAddress("$.opChainDeployments[0].l1StandardBridgeProxyAddress"));
        save(
            "L1CrossDomainMessengerProxy",
            _stateJson.readAddress("$.opChainDeployments[0].l1CrossDomainMessengerProxyAddress")
        );
        save(
            "DisputeGameFactoryProxy", _stateJson.readAddress("$.opChainDeployments[0].disputeGameFactoryProxyAddress")
        );
        save(
            "PermissionedDelayedWETHProxy",
            _stateJson.readAddress("$.opChainDeployments[0].delayedWETHPermissionedGameProxyAddress")
        );
        save(
            "AnchorStateRegistryProxy",
            _stateJson.readAddress("$.opChainDeployments[0].anchorStateRegistryProxyAddress")
        );
        save("AnchorStateRegistry", _stateJson.readAddress("$.opChainDeployments[0].anchorStateRegistryImplAddress"));
        save(
            "PermissionedDisputeGame", _stateJson.readAddress("$.opChainDeployments[0].permissionedDisputeGameAddress")
        );
        save("OptimismPortalProxy", _stateJson.readAddress("$.opChainDeployments[0].optimismPortalProxyAddress"));
    }
}
