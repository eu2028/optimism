// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import {CommonBase} from "../../lib/forge-std/src/Base.sol";
import {Script} from "../../lib/forge-std/src/Script.sol";
import {StdChains} from "../../lib/forge-std/src/StdChains.sol";
import {StdCheatsSafe} from "../../lib/forge-std/src/StdCheats.sol";
import {StdUtils} from "../../lib/forge-std/src/StdUtils.sol";
import {console} from "../../lib/forge-std/src/console.sol";
import {SuperchainConfig} from "../../src/L1/SuperchainConfig.sol";
import {FaultDisputeGame} from "../../src/dispute/FaultDisputeGame.sol";
import {PermissionedDisputeGame} from "../../src/dispute/PermissionedDisputeGame.sol";
import {IAnchorStateRegistry} from "../../src/dispute/interfaces/IAnchorStateRegistry.sol";
import {IBigStepper} from "../../src/dispute/interfaces/IBigStepper.sol";
import {IDelayedWETH} from "../../src/dispute/interfaces/IDelayedWETH.sol";
import {DelayedWETH} from "../../src/dispute/weth/DelayedWETH.sol";
import {Proxy} from "../../src/universal/Proxy.sol";
import {GameType, Duration, Claim} from "../../src/dispute/lib/LibUDT.sol";
import {GameTypes} from "../../src/dispute/lib/Types.sol";
import {DisputeGameFactory} from "../../src/dispute/DisputeGameFactory.sol";
import {Constants} from "../../src/libraries/Constants.sol";
import {ChainAssertions} from "./ChainAssertions.sol";

contract DeployPermissionless is Script {

    address public _disputeGameFactoryProxy = vm.envAddress("DGF_FACTORY_PROXY");
    bytes32 public _absolutePrestate = vm.envBytes32("ABSOLUTE_PRESTATE");


    function run() external {
        DisputeGameFactory dgf = DisputeGameFactory(_disputeGameFactoryProxy);

        PermissionedDisputeGame permissioned = PermissionedDisputeGame(payable(address(dgf.gameImpls(GameTypes.PERMISSIONED_CANNON))));
        Proxy proxy = deployDelayedWethProxy();
        DelayedWETH permDelayedWeth = DelayedWETH(payable(address(permissioned.weth())));
        initializeDelayedWethProxy(permDelayedWeth, proxy);
        transferDelayedWethProxyAdmin(permDelayedWeth, proxy);

        FaultDisputeGame fdg = deployFaultDisputeGame(permissioned, IDelayedWETH(address(proxy)));
        PermissionedDisputeGame pdg = deployPermissionedDisputeGame(permissioned);

        checkDelayedWeth(permDelayedWeth, DelayedWETH(payable(address(proxy))));
        checkFaultDisputeGame(permissioned, fdg, IDelayedWETH(address(proxy)));
        checkPermissionedDisputeGame(permissioned, pdg);
        printDeploymentSummary(address(proxy), address(fdg), address(pdg));
    }

    function deployDelayedWethProxy() internal broadcast returns (Proxy) {
        console.log(string.concat("Deploying ERC1967 proxy for DelayedWETH"));
        Proxy proxy = new Proxy({_admin: msg.sender});
        return proxy;
    }

    function initializeDelayedWethProxy(DelayedWETH _permissioned, Proxy _proxy) internal broadcast {
        console.log("Initializing proxy for DelayedWETH");
        address delayedWEthImpl = address(uint160(uint256(vm.load(address(_permissioned), Constants.PROXY_IMPLEMENTATION_ADDRESS))));
        _proxy.upgradeToAndCall(delayedWEthImpl, abi.encodeCall(DelayedWETH.initialize, (_permissioned.owner(), _permissioned.config())));
    }

    function transferDelayedWethProxyAdmin(DelayedWETH _permissioned, Proxy _proxy) internal broadcast {
        address proxyAdmin = address(uint160(uint256(vm.load(address(_permissioned), Constants.PROXY_OWNER_ADDRESS))));
        console.log("Transferring DelayedWETH proxy admin to ", proxyAdmin);
        _proxy.changeAdmin(proxyAdmin);
    }

    function deployFaultDisputeGame(PermissionedDisputeGame _permissioned, IDelayedWETH _delayedWeth) internal broadcast returns (FaultDisputeGame) {
        console.log("Deploying FaultDisputeGame");

        FaultDisputeGame impl = new FaultDisputeGame(
            GameTypes.CANNON,
            Claim.wrap(_absolutePrestate),
            _permissioned.maxGameDepth(),
            _permissioned.splitDepth(),
            _permissioned.clockExtension(),
            _permissioned.maxClockDuration(),
            _permissioned.vm(),
            _delayedWeth,
            _permissioned.anchorStateRegistry(),
            _permissioned.l2ChainId()
        );
        return impl;
    }

    function deployPermissionedDisputeGame(PermissionedDisputeGame _permissioned) internal broadcast returns (PermissionedDisputeGame) {
        console.log("Deploying PermissionedDisputeGame");
        PermissionedDisputeGame impl = new PermissionedDisputeGame(
            GameTypes.PERMISSIONED_CANNON,
            Claim.wrap(_absolutePrestate),
            _permissioned.maxGameDepth(),
            _permissioned.splitDepth(),
            _permissioned.clockExtension(),
            _permissioned.maxClockDuration(),
            _permissioned.vm(),
            _permissioned.weth(),
            _permissioned.anchorStateRegistry(),
            _permissioned.l2ChainId(),
            _permissioned.proposer(),
            _permissioned.challenger()
        );
        return impl;
    }

    function checkDelayedWeth(DelayedWETH _expected, DelayedWETH _impl) internal view {
        require(_impl.owner() == _expected.owner(), "WETH-10");
        require(_impl.delay() == _expected.delay(), "WETH-20");
        require(_impl.config() == _expected.config(), "WETH-30");
    }

    function checkFaultDisputeGame(PermissionedDisputeGame _permissioned, FaultDisputeGame _impl, IDelayedWETH _delayedWeth) internal view {
        require(_impl.gameType().raw() == GameTypes.CANNON.raw(), "FDG-10");
        require(_impl.absolutePrestate().raw() == _absolutePrestate, "FDG-20");
        require(_impl.maxGameDepth() == _permissioned.maxGameDepth(), "FDG-30");
        require(_impl.splitDepth() == _permissioned.splitDepth(), "FDG-40");
        require(_impl.clockExtension().raw() == _permissioned.clockExtension().raw(), "FDG-50");
        require(_impl.maxClockDuration().raw() == _permissioned.maxClockDuration().raw(), "FDG-60");
        require(_impl.vm() == _permissioned.vm(), "FDG-70");
        require(_impl.weth() == _delayedWeth, "FDG-80");
        require(_impl.anchorStateRegistry() == _permissioned.anchorStateRegistry(), "FDG-90");
        require(_impl.l2ChainId() == _permissioned.l2ChainId(), "FDG-100");
    }

    function checkPermissionedDisputeGame(PermissionedDisputeGame _permissioned, PermissionedDisputeGame _impl) internal view {
        require(_impl.gameType().raw() == GameTypes.PERMISSIONED_CANNON.raw(), "PDG-10");
        require(_impl.absolutePrestate().raw() == _absolutePrestate, "PDG-20");
        require(_impl.maxGameDepth() == _permissioned.maxGameDepth(), "PDG-30");
        require(_impl.splitDepth() == _permissioned.splitDepth(), "PDG-40");
        require(_impl.clockExtension().raw() == _permissioned.clockExtension().raw(), "PDG-50");
        require(_impl.maxClockDuration().raw() == _permissioned.maxClockDuration().raw(), "PDG-60");
        require(_impl.vm() == _permissioned.vm(), "PDG-70");
        require(_impl.weth() == _permissioned.weth(), "PDG-80");
        require(_impl.anchorStateRegistry() == _permissioned.anchorStateRegistry(), "PDG-90");
        require(_impl.l2ChainId() == _permissioned.l2ChainId(), "PDG-100");
    }

    /// @notice Prints a summary of the contracts deployed during this script.
    function printDeploymentSummary(address _wethProxy, address _fdg, address _pdg) internal view {
        console.log("Deployment Summary (chainid: %d)", block.chainid);
        console.log("    0. DelayedWETHProxy: %s", _wethProxy);
        console.log("    1. FaultDisputegame: %s", _fdg);
        console.log("    2. PermissionedDisputeGame: %s", _pdg);
    }

    ////////////////////////////////////////////////////////////////
    //                        Modifiers                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Modifier that wraps a function in broadcasting.
    modifier broadcast() {
        vm.startBroadcast(msg.sender);
        _;
        vm.stopBroadcast();
    }
}
