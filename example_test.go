package main

import (
	"fmt"
)

// Example demonstrates how to use the generated pool functions
func Example_usage() {
	// This is a demonstration of how the generated code would be used
	// After running the plugin on a .proto file, users would be able to:
	//
	// msg := GetMyMessage()  // Get message from pool
	// // ... use msg ...
	// PutMyMessage(msg)      // Return message to pool
	//
	// The actual functions would be generated based on the message names in the .proto file
	fmt.Println("Example of how to use generated pool functions")
}
