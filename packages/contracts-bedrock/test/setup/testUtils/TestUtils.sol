// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Vm } from "forge-std/Vm.sol";
import {
    IAddressCondition,
    AddressConditionChainer,
    AddressConditions,
    IsPayableCondition,
    IsNotPayableCondition,
    IsNotPrecompileCondition,
    IsNotForgeAddressCondition
} from "test/setup/testUtils/conditions/AddressConditions.sol";
import { ForbiddenAddresses, ForbiddenUint256 } from "test/setup/testUtils/Forbiddens.sol";

contract TestUtils {
    Vm private constant vm = Vm(address(bytes20(uint160(uint256(keccak256("hevm cheat code"))))));

    // This function returns addr if it satisfies all given conditions and is not forbidden,
    // otherwise it will generate a new random address that satisfies the conditions and is not
    // forbidden.
    // NOTE: This function will resort to vm.assume() if it does not find a valid address within
    // the given number of attempts.
    function _randomAddress(
        address addr,
        AddressConditionChainer _conditions,
        ForbiddenAddresses _forbiddenAddresses,
        uint256 attempts
    )
        internal
        returns (address)
    {
        bool pass = false;
        IAddressCondition[] memory conditions = _conditions.conditions();

        for (uint256 i; i < attempts; i++) {
            pass = true;
            if (_forbiddenAddresses.forbiddenAddresses(addr)) continue;
            for (uint256 j; j < conditions.length; j++) {
                if (!conditions[j].check(addr)) {
                    pass = false;
                    break;
                }
            }
            if (pass) break;
            addr = _randomAddress();
        }
        vm.assume(pass);

        return addr;
    }

    // This function returns addr if it is not forbidden by _forbiddenAddresses, otherwise it
    // will generate a new random address that is not forbidden.
    // NOTE: This function will resort to vm.assume() if it does not find a valid address within
    // the given number of attempts.
    function _randomAddress(
        address addr,
        ForbiddenAddresses _forbiddenAddresses,
        uint256 attempts
    )
        internal
        returns (address)
    {
        bool pass = false;

        for (uint256 i; i < attempts; i++) {
            if (_forbiddenAddresses.forbiddenAddresses(addr)) {
                pass = false;
                addr = _randomAddress();
            } else {
                pass = true;
            }
            if (pass) break;
        }
        vm.assume(pass);

        return addr;
    }

    // This function returns addr if it satisfies all given conditions, otherwise it will generate
    // a new random address that satisfies the conditions.
    // NOTE: This function will resort to vm.assume() if it does not find a valid address within
    // the given number of attempts.
    function _randomAddress(
        address addr,
        AddressConditionChainer _conditions,
        uint256 attempts
    )
        internal
        returns (address)
    {
        bool pass = false;
        IAddressCondition[] memory conditions = _conditions.conditions();

        for (uint256 i; i < attempts; i++) {
            pass = true;
            for (uint256 j; j < conditions.length; j++) {
                if (!conditions[j].check(addr)) {
                    pass = false;
                    break;
                }
            }
            if (pass) break;
            addr = _randomAddress();
        }
        vm.assume(pass);

        return addr;
    }

    function _randomAddress() internal returns (address) {
        return address(uint160(vm.randomUint()));
    }

    // This function returns _bound(_value, _min, _max) unless _bound(_value, _min, _max) is
    // forbidden by _forbiddenUint256, in which case it will generate a new random uint256 that
    // is not forbidden in the _forbiddenUint256 contract.
    // NOTE: This function will resort to vm.assume() if it does not find a valid uint256 within
    // the given number of attempts.
    function _boundExcept(
        uint256 _value,
        uint256 _min,
        uint256 _max,
        ForbiddenUint256 _forbiddenUint256,
        uint256 _attempts
    )
        internal
        returns (uint256)
    {
        uint256 value_ = __bound(_value, _min, _max);
        bool pass = false;
        for (uint256 i; i < _attempts; i++) {
            if (_forbiddenUint256.forbiddenUint256(value_)) {
                pass = false;
                value_ = __bound(vm.randomUint(), _min, _max);
            } else {
                pass = true;
            }
            if (pass) break;
        }
        vm.assume(pass);
        return value_;
    }

    function __bound(uint256 _value, uint256 _min, uint256 _max) private pure returns (uint256 value_) {
        value_ = (_value % (_max - _min)) + _min;
    }

    function __randomBytes(uint256 _minLength, uint256 _maxLength) internal returns (bytes memory bytes_) {
        uint256 length = __bound(vm.randomUint(), _minLength, _maxLength);
        bytes_ = new bytes(length);
        for (uint256 i; i < length; i++) {
            bytes_[i] = bytes1(uint8(vm.randomUint()));
        }
    }

    uint256 private isPayableConditionCounter;

    function newIsPayableCondition() internal returns (IsPayableCondition) {
        IsPayableCondition condition = new IsPayableCondition();
        vm.label(address(condition), string.concat("IsPayableCondition:", vm.toString(isPayableConditionCounter++)));
        return condition;
    }

    uint256 private isNotPayableConditionCounter;

    function newIsNotPayableCondition() internal returns (IsNotPayableCondition) {
        IsNotPayableCondition condition = new IsNotPayableCondition();
        vm.label(
            address(condition), string.concat("IsNotPayableCondition:", vm.toString(isNotPayableConditionCounter++))
        );
        return condition;
    }

    uint256 private isNotPrecompileConditionCounter;

    function newIsNotPrecompileCondition() internal returns (IsNotPrecompileCondition) {
        IsNotPrecompileCondition condition = new IsNotPrecompileCondition();
        vm.label(
            address(condition),
            string.concat("IsNotPrecompileCondition:", vm.toString(isNotPrecompileConditionCounter++))
        );
        return condition;
    }

    uint256 private isNotForgeAddressConditionCounter;

    function newIsNotForgeAddressCondition() internal returns (IsNotForgeAddressCondition) {
        IsNotForgeAddressCondition condition = new IsNotForgeAddressCondition();
        vm.label(
            address(condition),
            string.concat("IsNotForgeAddressCondition:", vm.toString(isNotForgeAddressConditionCounter++))
        );
        return condition;
    }

    uint256 private conditionsCounter;

    function newAddressConditionChainer() internal returns (AddressConditionChainer) {
        AddressConditionChainer conditions = new AddressConditionChainer();
        vm.label(address(conditions), string.concat("AddressConditionChainer:", vm.toString(conditionsCounter++)));
        return conditions;
    }

    uint256 private forbiddenAddressesCounter;

    function newForbiddenAddresses() internal returns (ForbiddenAddresses) {
        ForbiddenAddresses forbiddenAddresses = new ForbiddenAddresses();
        vm.label(
            address(forbiddenAddresses), string.concat("ForbiddenAddresses:", vm.toString(forbiddenAddressesCounter++))
        );
        return forbiddenAddresses;
    }

    uint256 private forbiddenUint256Counter;

    function newForbiddenUint256() internal returns (ForbiddenUint256) {
        ForbiddenUint256 forbiddenUint256 = new ForbiddenUint256();
        vm.label(address(forbiddenUint256), string.concat("ForbiddenUint256:", vm.toString(forbiddenUint256Counter++)));
        return forbiddenUint256;
    }
}
