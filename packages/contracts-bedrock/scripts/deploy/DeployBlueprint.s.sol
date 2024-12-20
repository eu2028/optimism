pragma solidity 0.8.15;

import { console2 as console } from "forge-std/console2.sol";

import { Script } from "forge-std/Script.sol";
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";
import { Blueprint } from "src/libraries/Blueprint.sol";
import { Bytes } from "src/libraries/Bytes.sol";

contract DeployBlueprintInput is BaseDeployIO {
    string internal _contractName;
    bool internal _big;
    bytes32 internal _salt;

    function set(bytes4 _sel, string memory _value) public {
        if (_sel == this.contractName.selector) {
            _contractName = _value;
        } else {
            revert("DeployBlueprint: unknown selector");
        }
    }

    function set(bytes4 _sel, bool _value) public {
        if (_sel == this.big.selector) {
            _big = _value;
        } else {
            revert("DeployBlueprint: unknown selector");
        }
    }

    function set(bytes4 _sel, bytes32 _value) public {
        if (_sel == this.salt.selector) {
            _salt = _value;
        } else {
            revert("DeployBlueprint: unknown selector");
        }
    }

    function contractName() public view returns (string memory) {
        return _contractName;
    }

    function big() public view returns (bool) {
        return _big;
    }

    function salt() public view returns (bytes32) {
        return _salt;
    }
}

contract DeployBlueprintOutput is BaseDeployIO {
    address internal _part0;
    address internal _part1;

    function set(bytes4 _sel, address _value) public {
        if (_sel == this.part0.selector) {
            require(_value != address(0), "DeployBlueprint: part0 cannot cannot be zero address");
            _part0 = _value;
        } else if (_sel == this.part1.selector) {
            _part1 = _value;
        } else {
            revert("DeployBlueprint: unknown selector");
        }
    }

    function part0() public view returns (address) {
        return _part0;
    }

    function part1() public view returns (address) {
        return _part1;
    }
}

contract DeployBlueprint is Script {
    function run(DeployBlueprintInput _dbi, DeployBlueprintOutput _dbo) public {
        address part0;
        address part1;
        bytes32 salt = _dbi.salt();
        string memory contractName = _dbi.contractName();

        if (_dbi.big()) {
            bytes memory bytecode = vm.getCode(contractName);
            (part0, part1) = deployBigBytecode(bytecode, salt);
        } else {
            bytes memory bytecode = Blueprint.blueprintDeployerBytecode(vm.getCode(contractName));
            vm.broadcast(msg.sender);
            part0 = deployBytecode(bytecode, salt);
        }

        _dbo.set(_dbo.part0.selector, part0);
        _dbo.set(_dbo.part1.selector, part1);
    }

    function deployPDG() public {
        DeployBlueprintInput dbi = new DeployBlueprintInput();
        DeployBlueprintOutput dbo = new DeployBlueprintOutput();
        bytes32 salt = bytes32(keccak256(bytes("supersalt")));
        string memory name = "PermissionedDisputeGame";

        dbi.set(dbi.contractName.selector, name);
        dbi.set(dbi.big.selector, true);
        dbi.set(dbi.salt.selector, salt);
        run(dbi, dbo);

        console.log("PDG:");
        console.log("part0:", dbo.part0());
        console.log("part1:", dbo.part1());
    }

    function deployASR() public {
        DeployBlueprintInput dbi = new DeployBlueprintInput();
        DeployBlueprintOutput dbo = new DeployBlueprintOutput();
        bytes32 salt = bytes32(keccak256(bytes("supersalt")));
        string memory name = "AnchorStateRegistry";

        dbi.set(dbi.contractName.selector, name);
        dbi.set(dbi.big.selector, false);
        dbi.set(dbi.salt.selector, salt);
        run(dbi, dbo);

        console.log("ASR:");
        console.log("part0:", dbo.part0());
        console.log("part1:", dbo.part1());
    }

    function deployBytecode(bytes memory _bytecode, bytes32 _salt) public returns (address newContract_) {
        assembly ("memory-safe") {
            newContract_ := create2(0, add(_bytecode, 0x20), mload(_bytecode), _salt)
        }
        require(newContract_ != address(0), "DeployBlueprint: create2 failed");
    }

    function deployBigBytecode(
        bytes memory _bytecode,
        bytes32 _salt
    )
        public
        returns (address newContract1_, address newContract2_)
    {
        // Preamble needs 3 bytes.
        uint256 maxInitCodeSize = 24576 - 3;
        require(_bytecode.length > maxInitCodeSize, "DeployBlueprint: Use deployBytecode instead");

        bytes memory part1Slice = Bytes.slice(_bytecode, 0, maxInitCodeSize);
        bytes memory part1 = Blueprint.blueprintDeployerBytecode(part1Slice);
        bytes memory part2Slice = Bytes.slice(_bytecode, maxInitCodeSize, _bytecode.length - maxInitCodeSize);
        bytes memory part2 = Blueprint.blueprintDeployerBytecode(part2Slice);

        vm.startBroadcast(msg.sender);
        newContract1_ = deployBytecode(part1, _salt);
        newContract2_ = deployBytecode(part2, _salt);
        vm.stopBroadcast();
    }
}
