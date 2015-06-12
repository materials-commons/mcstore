package mccli

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

const maxSimultaneous = 5

// getNumThreads ensures that the number of parallel downloads is valid.
func getNumThreads(c *cli.Context) int {
	numThreads := c.Int("parallel")

	if numThreads < 1 {
		fmt.Println("Simultaneous downloads must be positive: ", numThreads)
		os.Exit(1)
	} else if numThreads > maxSimultaneous {
		fmt.Printf("You may not set simultaneous downloads greater than %d: %d\n", maxSimultaneous, numThreads)
		os.Exit(1)
	}

	return numThreads
}
