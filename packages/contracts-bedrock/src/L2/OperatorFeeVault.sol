// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { FeeVault } from "src/L2/FeeVault.sol";

// Libraries
import { Types } from "src/libraries/Types.sol";

// Interfaces
import { ISemver } from "interfaces/universal/ISemver.sol";

/// @custom:proxied true
/// @custom:predeploy 0x420000000000000000000000000000000000001B
/// @title OperatorFeeVault
/// @notice The OperatorFeeVault accumulates the operator portion of the transaction fees.
contract OperatorFeeVault is FeeVault, ISemver {
    /// @notice Semantic version.
    /// @custom:semver 1.5.0-beta.5
    string public constant version = "1.5.0-beta.5";

    /// @notice Constructs the OperatorFeeVault contract.
    /// @param _recipient           Wallet that will receive the fees.
    /// @param _minWithdrawalAmount Minimum balance for withdrawals.
    /// @param _withdrawalNetwork   Network which the recipient will receive fees on.
    constructor(
        address _recipient,
        uint256 _minWithdrawalAmount,
        Types.WithdrawalNetwork _withdrawalNetwork
    )
        FeeVault(_recipient, _minWithdrawalAmount, _withdrawalNetwork)
    { }
}
