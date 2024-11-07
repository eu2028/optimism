// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing utilities
import { CommonTest } from "test/setup/CommonTest.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Types } from "src/libraries/Types.sol";
import "src/libraries/PortalErrors.sol";

// Target contract dependencies
import "src/libraries/PortalErrors.sol";
import { OptimismPortalInterop } from "src/L1/OptimismPortalInterop.sol";
import { L1BlockInterop } from "src/L2/L1BlockInterop.sol";

// Interfaces
import { IOptimismPortalInterop } from "src/L1/interfaces/IOptimismPortalInterop.sol";

contract OptimismPortalInterop_Test is CommonTest {
    /// @notice Marked virtual to be overridden in
    ///         test/kontrol/deployment/DeploymentSummary.t.sol
    function setUp() public virtual override {
        super.enableInterop();
        super.setUp();
    }

    /// @dev Tests that any config type can be set by the system config.
    function testFuzz_setConfig_succeeds(uint8 _configType, bytes calldata _value) public {
        // Ensure that _configType is within the range of the ConfigType enum
        _configType = uint8(bound(uint256(_configType), 0, uint256(type(Types.ConfigType).max)));

        vm.expectEmit(address(optimismPortal));
        emitTransactionDeposited({
            _from: Constants.DEPOSITOR_ACCOUNT,
            _to: Predeploys.L1_BLOCK_ATTRIBUTES,
            _value: 0,
            _mint: 0,
            _gasLimit: 200_000,
            _isCreation: false,
            _data: abi.encodeCall(L1BlockInterop.setConfig, (Types.ConfigType(_configType), _value))
        });

        vm.prank(address(_optimismPortalInterop().systemConfig()));
        _optimismPortalInterop().setConfig(Types.ConfigType(_configType), _value);
    }

    /// @dev Tests that setting the gas paying token config as not the system config reverts.
    function testFuzz_setConfig_gasPayingTokenButNotSystemConfig_reverts(bytes calldata _value) public {
        vm.expectRevert(Unauthorized.selector);
        _optimismPortalInterop().setConfig(Types.ConfigType.GAS_PAYING_TOKEN, _value);
    }

    /// @dev Tests that the config for adding a dependency can be set.
    function testFuzz_setConfig_addDependency_succeeds(bytes calldata _value) public {
        vm.expectEmit(address(optimismPortal));
        emitTransactionDeposited({
            _from: Constants.DEPOSITOR_ACCOUNT,
            _to: Predeploys.L1_BLOCK_ATTRIBUTES,
            _value: 0,
            _mint: 0,
            _gasLimit: 200_000,
            _isCreation: false,
            _data: abi.encodeCall(L1BlockInterop.setConfig, (Types.ConfigType.ADD_DEPENDENCY, _value))
        });

        vm.prank(address(_optimismPortalInterop().systemConfig()));
        _optimismPortalInterop().setConfig(Types.ConfigType.ADD_DEPENDENCY, _value);
    }

    /// @dev Tests that setting the add dependency config as not the system config reverts.
    function testFuzz_setConfig_addDependencyButNotSystemConfig_reverts(bytes calldata _value) public {
        vm.expectRevert(Unauthorized.selector);
        _optimismPortalInterop().setConfig(Types.ConfigType.ADD_DEPENDENCY, _value);
    }

    /// @dev Tests that the config for removing a dependency can be set.
    function testFuzz_setConfig_removeDependency_succeeds(bytes calldata _value) public {
        vm.expectEmit(address(optimismPortal));
        emitTransactionDeposited({
            _from: Constants.DEPOSITOR_ACCOUNT,
            _to: Predeploys.L1_BLOCK_ATTRIBUTES,
            _value: 0,
            _mint: 0,
            _gasLimit: 200_000,
            _isCreation: false,
            _data: abi.encodeCall(L1BlockInterop.setConfig, (Types.ConfigType.REMOVE_DEPENDENCY, _value))
        });

        vm.prank(address(_optimismPortalInterop().systemConfig()));
        _optimismPortalInterop().setConfig(Types.ConfigType.REMOVE_DEPENDENCY, _value);
    }

    /// @dev Tests that setting the remove dependency config as not the system config reverts.
    function testFuzz_setConfig_removeDependencyButNotSystemConfig_reverts(bytes calldata _value) public {
        vm.expectRevert(Unauthorized.selector);
        _optimismPortalInterop().setConfig(Types.ConfigType.REMOVE_DEPENDENCY, _value);
    }

    /// @dev Tests that setting any config type as not the system config reverts.
    function testFuzz_setConfig_notSystemConfig_reverts(
        address _caller,
        uint8 _configType,
        bytes calldata _value
    )
        external
    {
        // Ensure that _configType is within the range of the Types.ConfigType enum
        _configType = uint8(bound(uint256(_configType), 0, uint256(type(Types.ConfigType).max)));

        vm.assume(_caller != address(_optimismPortalInterop().systemConfig()));
        vm.prank(_caller);
        vm.expectRevert(Unauthorized.selector);
        _optimismPortalInterop().setConfig(Types.ConfigType(_configType), _value);
    }

    /// @dev Returns the OptimismPortalInterop instance.
    function _optimismPortalInterop() internal view returns (IOptimismPortalInterop) {
        return IOptimismPortalInterop(payable(address(optimismPortal)));
    }
}
