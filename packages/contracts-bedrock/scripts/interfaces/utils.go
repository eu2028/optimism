package main

func isTrivialType(typeString string) bool {
	trivialTypes := []string{"uint256", "int256", "bool", "address", "bytes32", "uint", "int"}

	for _, t := range trivialTypes {
		if typeString == t {
			return true
		}
	}

	return false
}
