// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Ephemeral contract that stores a list of forbidden addresses.
contract ForbiddenAddresses {
    mapping(address => bool) public forbiddenAddresses;

    // chainable
    function forbid(address _addr) public returns (ForbiddenAddresses) {
        forbiddenAddresses[_addr] = true;
        return this;
    }
}

// Ephemeral contract that stores a list of forbidden uint256 values.
contract ForbiddenUint256 {
    mapping(uint256 => bool) public forbiddenUint256;

    // chainable
    function forbid(uint256 _value) public returns (ForbiddenUint256) {
        forbiddenUint256[_value] = true;
        return this;
    }
}
