package denominations

import "fmt"

// ITC Units - Base denomination constants
const (
	Ticks   = 1    // Smallest unit (1 Tick)
	Intello = 1e9  // 1 Intello = 1 billion Ticks
	ITC     = 1e18 // 1 ITC = 1 quintillion Ticks
)

// Conversion functions from larger to smaller units
func IntelloToTicks(intello uint64) uint64 {
	return intello * Intello
}

func ITCToTicks(itc uint64) uint64 {
	return itc * ITC
}

// Conversion functions from smaller to larger units
func TicksToIntello(ticks uint64) float64 {
	return float64(ticks) / float64(Intello)
}

func TicksToITC(ticks uint64) float64 {
	return float64(ticks) / float64(ITC)
}

// Additional helper functions for common operations
func ITCToIntello(itc uint64) uint64 {
	return itc * (ITC / Intello)
}

func IntelloToITC(intello uint64) float64 {
	return float64(intello) / float64(ITC/Intello)
}

// Function to format amounts with proper precision
func FormatTicks(ticks uint64) string {
	itc := TicksToITC(ticks)
	return fmt.Sprintf("%.18f ITC", itc)
}
