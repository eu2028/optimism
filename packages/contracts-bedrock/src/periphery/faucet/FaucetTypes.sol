// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @notice Parameters for a drip.
struct DripParameters {
    address payable recipient;
    bytes data;
    bytes32 nonce;
}
