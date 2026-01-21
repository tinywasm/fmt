package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/tinywasm/fmt"
)

// BenchmarkResult stores benchmark results for memory analysis
type BenchmarkResult struct {
	Name        string
	Library     string
	Iterations  int64
	NsPerOp     int64
	BytesPerOp  int64
	AllocsPerOp int64
	Description string
}

// MemoryComparison stores comparison data between implementations
type MemoryComparison struct {
	Standard BenchmarkResult
	fmt      BenchmarkResult
	Category string
}

func main() {
	if len(os.Args) < 2 {
		println(Sprintf("Usage: go run analyzer.go [binary|memory|all]"))
		println(Sprintf("  binary  - Analyze binary sizes"))
		println(Sprintf("  memory  - Analyze memory allocations"))
		println(Sprintf("  all     - Run both analyses"))
		return
	}

	mode := os.Args[1]

	switch mode {
	case "binary":
		analyzeBinarySizes()
	case "memory":
		analyzeMemoryAllocations()
	case "all":
		analyzeBinarySizes()
		println()
		analyzeMemoryAllocations()
	default:
		LogError(Sprintf("Unknown mode: %s", mode))
		return
	}
}

// analyzeBinarySizes analyzes and reports binary size comparisons
func analyzeBinarySizes() {
	LogStep("Analyzing binary sizes with multiple optimization levels...")

	binaries := measureBinarySizes()
	if len(binaries) == 0 {
		LogError("No binaries found to analyze")
		return
	}

	displayBinaryResults(binaries)
	displayOptimizationTable(binaries)
	updateREADMEWithBinaryData(binaries)

	LogSuccess("Binary size analysis completed and README updated")
}

// analyzeMemoryAllocations analyzes and reports memory allocation comparisons
func analyzeMemoryAllocations() {
	LogStep("Starting memory allocation benchmark...")

	// Check if we can run benchmarks
	if !checkGoBenchAvailable() {
		LogError("Cannot run Go benchmarks")
		return
	}

	// Run memory benchmarks
	comparisons := runMemoryBenchmarks()
	if len(comparisons) == 0 {
		LogError("No benchmark results available. Make sure Go benchmarks can run successfully.")
		return
	}

	// Display results
	displayMemoryResults(comparisons)

	// Update README
	updateREADMEWithMemoryData(comparisons)

	LogSuccess("Memory benchmark completed and README updated")
}

// measureBinarySizes scans for and measures all binary files
func measureBinarySizes() []BinaryInfo {
	var allBinaries []BinaryInfo

	binaryDir := "bench-binary-size"
	if !FileExists(binaryDir) {
		LogError(Sprintf("Binary directory %s not found", binaryDir))
		return nil
	}

	// Define patterns to search for
	patterns := []string{"standard", "tinystring"}

	// Search for binaries
	for _, pattern := range patterns {
		binaries, err := FindBinaries(binaryDir, []string{pattern})
		if err != nil {
			LogError(Sprintf("Error finding binaries: %v", err))
			continue
		}
		allBinaries = append(allBinaries, binaries...)
	}

	return allBinaries
}

// displayBinaryResults shows binary size results in a table format
func displayBinaryResults(binaries []BinaryInfo) {
	println("\n📊 Binary Size Results:")
	println("========================")
	println(Sprintf("%-20s %-8s %-12s %-10s", "File", "Type", "Library", "Size"))
	println(Convert("-").Repeat(55).String())

	for _, binary := range binaries {
		println(Sprintf("%-20s %-8s %-12s %-10s",
			binary.Name, binary.Type, binary.Library, binary.SizeStr))
	}
	println("")
}

// displayOptimizationTable shows optimization comparison table
func displayOptimizationTable(binaries []BinaryInfo) {
	optimizations := getOptimizationConfigs()

	println("📊 Optimization Level Comparison:")
	println("==================================")

	for _, opt := range optimizations {
		println("")
		println(Sprintf("%s Optimization (%s):", opt.Name, opt.Description))
		println(Sprintf("%-15s %-15s %-15s %-15s", "", "Standard", "fmt", "Improvement"))
		println(Convert("-").Repeat(65).String())

		// Find matching binaries for this optimization level
		standardNative := findBinaryByPattern(binaries, "standard", "native", opt.Suffix)
		tinystringNative := findBinaryByPattern(binaries, "tinystring", "native", opt.Suffix)
		standardWasm := findBinaryByPattern(binaries, "standard", "wasm", opt.Suffix)
		tinystringWasm := findBinaryByPattern(binaries, "tinystring", "wasm", opt.Suffix)

		if standardNative.Name != "" && tinystringNative.Name != "" {
			improvement := calculateImprovement(standardNative.Size, tinystringNative.Size)
			println(Sprintf("%-15s %-15s %-15s %-15s", "Native",
				standardNative.SizeStr, tinystringNative.SizeStr, improvement))
		}

		if standardWasm.Name != "" && tinystringWasm.Name != "" {
			improvement := calculateImprovement(standardWasm.Size, tinystringWasm.Size)
			println(Sprintf("%-15s %-15s %-15s %-15s", "WebAssembly",
				standardWasm.SizeStr, tinystringWasm.SizeStr, improvement))
		}
	}
}

// findBinaryByPattern finds a binary matching the specified criteria
func findBinaryByPattern(binaries []BinaryInfo, library, binaryType, optSuffix string) BinaryInfo {
	for _, binary := range binaries {
		if binary.Library == library && binary.Type == binaryType && binary.OptLevel == extractOptLevel(binary.Name) {
			if optSuffix == "" && binary.OptLevel == "default" {
				return binary
			}
			if optSuffix != "" && Contains(binary.Name, optSuffix) {
				return binary
			}
		}
	}
	return BinaryInfo{}
}

// calculateImprovement calculates percentage improvement
func calculateImprovement(original, improved int64) string {
	if original == 0 {
		return "N/A"
	}

	improvement := float64(original-improved) / float64(original) * 100
	if improvement > 0 {
		return Sprintf("%.1f%% smaller", improvement)
	} else if improvement < 0 {
		return Sprintf("%.1f%% larger", -improvement)
	}
	return "Same size"
}

// getOptimizationConfigs returns TinyGo optimization configurations
func getOptimizationConfigs() []OptimizationConfig {
	return []OptimizationConfig{
		{
			Name:        "Default",
			Flags:       "",
			Description: "Default TinyGo optimization (-opt=z)",
			Suffix:      "",
		},
		{
			Name:        "Ultra",
			Flags:       "-opt=z -gc=leaking -scheduler=none",
			Description: "Ultra size optimization",
			Suffix:      "-ultra",
		},
		{
			Name:        "Speed",
			Flags:       "-opt=2",
			Description: "Speed optimization",
			Suffix:      "-speed",
		},
		{
			Name:        "Debug",
			Flags:       "-opt=1",
			Description: "Debug build",
			Suffix:      "-debug",
		},
	}
}

// checkGoBenchAvailable checks if Go benchmarks can be run
func checkGoBenchAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// runMemoryBenchmarks executes memory benchmarks and returns comparisons
func runMemoryBenchmarks() []MemoryComparison {
	var comparisons []MemoryComparison

	// Run standard library benchmarks
	LogInfo("Running standard library memory benchmarks...")
	standardResults := runBenchmarks("standard")

	// Run fmt benchmarks
	LogInfo("Running fmt memory benchmarks...")
	tinystringResults := runBenchmarks("tinystring")

	// Create comparisons
	comparisons = append(comparisons, createComparison(
		"String Processing",
		findBenchmark(standardResults, "BenchmarkStringProcessing"),
		findBenchmark(tinystringResults, "BenchmarkStringProcessing"),
	))

	comparisons = append(comparisons, createComparison(
		"Number Processing",
		findBenchmark(standardResults, "BenchmarkNumberProcessing"),
		findBenchmark(tinystringResults, "BenchmarkNumberProcessing"),
	))

	comparisons = append(comparisons, createComparison(
		"Mixed Operations",
		findBenchmark(standardResults, "BenchmarkMixedOperations"),
		findBenchmark(tinystringResults, "BenchmarkMixedOperations"),
	))

	// Check for pointer optimization benchmark (fmt only)
	pointerBench := findBenchmark(tinystringResults, "BenchmarkStringProcessingWithPointers")
	if pointerBench.Name != "" {
		standardEquivalent := findBenchmark(standardResults, "BenchmarkStringProcessing")
		comparisons = append(comparisons, createComparison(
			"String Processing (Pointer Optimization)",
			standardEquivalent,
			pointerBench,
		))
	}

	return comparisons
}

// runBenchmarks executes benchmarks for a specific library implementation
func runBenchmarks(library string) []BenchmarkResult {
	var results []BenchmarkResult

	benchDir := filepath.Join("bench-memory-alloc", library)
	if !FileExists(benchDir) {
		LogError(Sprintf("Benchmark directory %s not found", benchDir))
		return results
	}
	cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-run=^$")
	cmd.Dir = benchDir

	output, err := cmd.Output()
	if err != nil {
		LogError(Sprintf("Failed to run benchmarks in %s: %v", benchDir, err))
		return results
	}

	return parseBenchmarkOutput(string(output), library)
}

// parseBenchmarkOutput parses Go benchmark output into structured results
func parseBenchmarkOutput(output, library string) []BenchmarkResult {
	var results []BenchmarkResult

	scanner := bufio.NewScanner(strings.NewReader(output))
	benchmarkRegex := regexp.MustCompile(`^(Benchmark\w+)(?:-\d+)?\s+(\d+)\s+(\d+)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)
	for scanner.Scan() {
		line := scanner.Text()
		matches := benchmarkRegex.FindStringSubmatch(line)

		if len(matches) == 6 {
			iterations, _ := Convert(matches[2]).Int64()
			nsPerOp, _ := Convert(matches[3]).Int64()
			bytesPerOp, _ := Convert(matches[4]).Int64()
			allocsPerOp, _ := Convert(matches[5]).Int64()

			out := BenchmarkResult{
				Name:        matches[1],
				Library:     library,
				Iterations:  iterations,
				NsPerOp:     nsPerOp,
				BytesPerOp:  bytesPerOp,
				AllocsPerOp: allocsPerOp,
			}

			results = append(results, out)
		}
	}

	return results
}

// createComparison creates a memory comparison between two benchmark results
func createComparison(category string, standard, tinystring BenchmarkResult) MemoryComparison {
	return MemoryComparison{
		Standard: standard,
		fmt:      tinystring,
		Category: category,
	}
}

// findBenchmark finds a benchmark out by name
func findBenchmark(results []BenchmarkResult, name string) BenchmarkResult {
	for _, out := range results {
		if out.Name == name {
			return out
		}
	}
	return BenchmarkResult{}
}

// displayMemoryResults shows memory benchmark results in a table format
func displayMemoryResults(comparisons []MemoryComparison) {
	println("\n🧠 Memory Allocation Results:")
	println("============================")
	println(Sprintf("%-35s %-12s %-15s %-15s %-15s",
		"Category", "Library", "Bytes/Op", "Allocs/Op", "Time/Op"))
	println(Convert("-").Repeat(95).String())

	for _, comparison := range comparisons {
		if comparison.Standard.Name != "" {
			println(Sprintf("%-35s %-12s %-15s %-15d %-15s",
				comparison.Category, "standard",
				FormatSize(comparison.Standard.BytesPerOp),
				comparison.Standard.AllocsPerOp,
				formatNanoTime(comparison.Standard.NsPerOp)))
		}

		if comparison.fmt.Name != "" {
			println(Sprintf("%-35s %-12s %-15s %-15d %-15s",
				"", "tinystring",
				FormatSize(comparison.fmt.BytesPerOp),
				comparison.fmt.AllocsPerOp,
				formatNanoTime(comparison.fmt.NsPerOp)))

			// Show improvement
			if comparison.Standard.Name != "" && comparison.fmt.Name != "" {
				memImprovement := calculateMemoryImprovement(
					comparison.Standard.BytesPerOp, comparison.fmt.BytesPerOp)
				allocImprovement := calculateMemoryImprovement(
					comparison.Standard.AllocsPerOp, comparison.fmt.AllocsPerOp)

				println(Sprintf("%-35s %-12s %-15s %-15s %-15s",
					"  → Improvement", "", memImprovement, allocImprovement, ""))
			}
		}
		println("")
	}
}

// updateREADMEWithBinaryData updates README with binary size analysis
func updateREADMEWithBinaryData(binaries []BinaryInfo) {
	reporter := NewReportGenerator("./README.md")
	if err := reporter.UpdateBinaryData(binaries); err != nil {
		LogError(Sprintf("Failed to update README with binary data: %v", err))
	}
}

// updateREADMEWithMemoryData updates README with memory benchmark data
func updateREADMEWithMemoryData(comparisons []MemoryComparison) {
	reporter := NewReportGenerator("./README.md")
	if err := reporter.UpdateMemoryData(comparisons); err != nil {
		LogError(Sprintf("Failed to update README with memory data: %v", err))
	}
}
