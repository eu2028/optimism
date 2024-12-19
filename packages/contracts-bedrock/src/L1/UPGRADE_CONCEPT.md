## Background

During upgrades we currently do the following two steps for all contracts being
upgraded:

1. reset the `_initialized` value by calling `upgradeToAndCall(StorageSetter,
setBool(slot, false)`. (the slot to delete differs per contract)
2. update the implementation and call `initialize()` by calling
   `upgradeToAndCall(implementation, initialize(lotsOfInputs)`

Note that if we're not trying to modify or add any state variables, there's no
need to call `initialize()`. In fact, AFAICT we could have just been using
`upgradeTo(implementation)` for most contracts in most upgrades we've done so
far.

## The current plan

The [current
proposal](https://github.com/ethereum-optimism/optimism/issues/13071) is to move
to unstructured storage for the initialized layout value.

However, there is a valid concern about changing the initializer storage slot,
especially when using a new upgrade process.

## Alternative

Instead maybe we can go the opposite direction and **stop touching the
initialized** value after a contract is deployed and move to a **single step
grade mechanism**.

## Design

We need to think about all the cases in the deploy/upgrade lifecycle of a contract.

### 1. Fresh deployment

These contracts have `upgradeToAndCall(implementation, intializeData)` called on them
by `opcm.deploy`. Nothing changes here, except for how the `initialize()` function
is extended in future upgrades.

### 2. A contract being upgraded with no new storage values being added

`opcm.upgrade()` will just call `upgradeTo(implementation)` on them. Simple and low risk.

### 3. A contract being upgraded with new storage values being added.

Using the SimpleStorage example included in this PR:

`opcm.upgrade()` will call
`upgradeToAndCall(implementation, abi.encodeCall(SimpleStorage.upgrade160(newValue)))`.
