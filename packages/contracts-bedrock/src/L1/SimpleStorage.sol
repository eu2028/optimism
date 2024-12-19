// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

// Libraries
import { Storage } from "src/libraries/Storage.sol";

contract SimpleStorage is Initializable {
    // Storage slots used to manage upgrades
    bytes32 internal constant UPGRADED_1_3_0 = keccak256("UPGRADED_1_3_0");
    bytes32 internal constant UPGRADED_1_6_0 = keccak256("UPGRADED_1_6_0");

    uint256 public value100;
    uint256 public value130;
    uint256 public value160;

    function initialize(uint256 _value100, uint256 _value130, uint256 _value160) external initializer {
        value100 = _value100;
        upgrade130(_value130);
        upgrade160(_value160);
    }

    function upgrade130(uint256 _value130) public {
        require(initialized, "SimpleStorage: must initialize first");
        require(!Storage.getBool(UPGRADED_1_3_0), "SimpleStorage: already upgraded to 1.3.0");
        value130 = _value130;
        Storage.setBool(UPGRADED_1_3_0, true);
    }

    function upgrade160(uint256 _value160) public {
        require(Storage.getBool(UPGRADED_1_3_0), "SimpleStorage: must upgrade to 1.3.0 first");
        require(!Storage.getBool(UPGRADED_1_6_0), "SimpleStorage: already upgraded to 1.6.0");
        value160 = _value160;
        Storage.setBool(UPGRADED_1_6_0, true);
    }
}
