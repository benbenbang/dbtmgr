package utils

import (
	"fmt"
	"sort"
)

func PrintTree(treeMap map[string]interface{}, prefix string) {
	// Get the keys and sort them for consistent output
	keys := make([]string, 0, len(treeMap))
	for key := range treeMap {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Sort the keys for alphabetical output

	// Iterate over the sorted keys and print
	for i, key := range keys {
		// Determine if this is the last element
		isLast := i == len(keys)-1

		// Print the current tree level
		fmt.Print(prefix)

		if isLast {
			fmt.Print("└── ")
		} else {
			fmt.Print("├── ")
		}

		// If the value is another map, it's a directory; otherwise, it's a file
		if subTree, isSubTree := treeMap[key].(map[string]interface{}); isSubTree {
			fmt.Println(key)
			// Create a new prefix for the next tree level
			newPrefix := prefix
			if isLast {
				newPrefix += "    " // Add spacing for alignment if it's the last element
			} else {
				newPrefix += "│   " // Add a vertical line for the next level
			}
			PrintTree(subTree, newPrefix) // Recurse into the sub-tree
		} else {
			fmt.Println(key) // Print the file name
		}
	}
}
