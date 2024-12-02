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
        console.log("DelayedWethProxy:", address(proxy));

        deployFaultDisputeGame(permissioned, IDelayedWETH(address(proxy)));
        deployPermissionedDisputeGame(permissioned);
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

    function deployFaultDisputeGame(PermissionedDisputeGame _permissioned, IDelayedWETH _delayedWeth) internal broadcast {
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
        console.log("FaultDisputeGame: ", address(impl));
    }

    function deployPermissionedDisputeGame(PermissionedDisputeGame _permissioned) internal broadcast {
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
        console.log("PermissionedDisputeGame: ", address(impl));
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
