package main

import (
	"fmt"
	"os"
	"strings" // Only for section finding in README
	"time"

	. "github.com/tinywasm/fmt"
)

// ReportGenerator handles README and documentation generation
type ReportGenerator struct {
	ReadmePath string
	TempPath   string
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(readmePath string) *ReportGenerator {
	return &ReportGenerator{
		ReadmePath: readmePath,
		TempPath:   readmePath + ".tmp",
	}
}

// UpdateREADMEWithBinaryData updates README with binary size comparison data
func (r *ReportGenerator) UpdateBinaryData(binaries []BinaryInfo) error {
	LogInfo("Updating README with binary size analysis...")

	content, err := r.generateBinarySizeSection(binaries)
	if err != nil {
		return Err(err)
	}

	return r.updateREADMESection("Binary Size Comparison", content)
}

// UpdateREADMEWithMemoryData updates README with memory benchmark data
func (r *ReportGenerator) UpdateMemoryData(comparisons []MemoryComparison) error {
	LogInfo("Updating README with memory allocation analysis...")

	content, err := r.generateMemorySection(comparisons)
	if err != nil {
		return Errf("failed to generate memory section: %v", err)
	}

	return r.updateREADMESection("Memory Usage Comparison", content)
}

// generateBinarySizeSection creates the binary size comparison section
func (r *ReportGenerator) generateBinarySizeSection(binaries []BinaryInfo) (string, error) {
	content := Convert() // Using fmt builder for better performance

	content.Write("## Binary Size Comparison\n\n")
	content.Write("[Standard Library Example](bench-binary-size/standard-lib/main.go) | [fmt Example](bench-binary-size/tinystring-lib/main.go)\n\n")
	content.Write("<!-- This table is automatically generated from build-and-measure.sh -->\n")
	content.Write("*Last updated: " + time.Now().Format("2006-01-02 15:04:05") + "*\n\n")

	// Group binaries by optimization level
	optimizations := getOptimizationConfigs()
	content.Write("| Build Type | Parameters | Standard Library<br/>`go build` | fmt<br/>`tinygo build` | Size Reduction | Performance |\n")
	content.Write("|------------|------------|------------------|------------|----------------|-------------|\n")

	var allImprovements []float64
	var maxImprovement float64
	var totalSavings int64

	for _, opt := range optimizations {
		// Find matching binaries for this optimization level
		standardNative := findBinaryByPattern(binaries, "standard", "native", opt.Suffix)
		tinystringNative := findBinaryByPattern(binaries, "tinystring", "native", opt.Suffix)
		standardWasm := findBinaryByPattern(binaries, "standard", "wasm", opt.Suffix)
		tinystringWasm := findBinaryByPattern(binaries, "tinystring", "wasm", opt.Suffix)

		// Build type icons and names
		buildIcon := getBuildTypeIcon(opt.Name)
		parameters := getBuildParameters(opt.Name, false)    // Native
		wasmParameters := getBuildParameters(opt.Name, true) // WASM
		// Native builds
		if standardNative.Name != "" && tinystringNative.Name != "" {
			improvementPercent := calculateImprovementPercent(standardNative.Size, tinystringNative.Size)
			sizeDiff := standardNative.Size - tinystringNative.Size
			performanceIndicator := getPerformanceIndicator(improvementPercent)

			content.Write(Fmt("| %s **%s Native** | `%s` | %s | %s | **-%s** | %s **%.1f%%** |\n",
				buildIcon, capitalizeFirst(opt.Name), parameters,
				standardNative.SizeStr, tinystringNative.SizeStr,
				FormatSize(sizeDiff), performanceIndicator, improvementPercent))

			allImprovements = append(allImprovements, improvementPercent)
			if improvementPercent > maxImprovement {
				maxImprovement = improvementPercent
			}
			totalSavings += sizeDiff
		}
		// WebAssembly builds
		if standardWasm.Name != "" && tinystringWasm.Name != "" {
			improvementPercent := calculateImprovementPercent(standardWasm.Size, tinystringWasm.Size)
			sizeDiff := standardWasm.Size - tinystringWasm.Size
			performanceIndicator := getPerformanceIndicator(improvementPercent)

			content.Write(Fmt("| 🌐 **%s WASM** | `%s` | %s | %s | **-%s** | %s **%.1f%%** |\n",
				capitalizeFirst(opt.Name), wasmParameters,
				standardWasm.SizeStr, tinystringWasm.SizeStr,
				FormatSize(sizeDiff), performanceIndicator, improvementPercent))

			allImprovements = append(allImprovements, improvementPercent)
			if improvementPercent > maxImprovement {
				maxImprovement = improvementPercent
			}
			totalSavings += sizeDiff
		}
	}

	// Calculate averages
	var avgImprovement float64
	var avgWasmImprovement float64
	var avgNativeImprovement float64
	var wasmCount, nativeCount int

	for i, opt := range optimizations {
		standardNative := findBinaryByPattern(binaries, "standard", "native", opt.Suffix)
		tinystringNative := findBinaryByPattern(binaries, "tinystring", "native", opt.Suffix)
		standardWasm := findBinaryByPattern(binaries, "standard", "wasm", opt.Suffix)
		tinystringWasm := findBinaryByPattern(binaries, "tinystring", "wasm", opt.Suffix)

		if standardNative.Name != "" && tinystringNative.Name != "" {
			improvement := calculateImprovementPercent(standardNative.Size, tinystringNative.Size)
			avgNativeImprovement += improvement
			nativeCount++
		}

		if standardWasm.Name != "" && tinystringWasm.Name != "" {
			improvement := calculateImprovementPercent(standardWasm.Size, tinystringWasm.Size)
			avgWasmImprovement += improvement
			wasmCount++
		}
		_ = i
	}

	if len(allImprovements) > 0 {
		for _, imp := range allImprovements {
			avgImprovement += imp
		}
		avgImprovement /= float64(len(allImprovements))
	}

	if nativeCount > 0 {
		avgNativeImprovement /= float64(nativeCount)
	}
	if wasmCount > 0 {
		avgWasmImprovement /= float64(wasmCount)
	}
	// Performance summary
	content.Write("\n### 🎯 Performance Summary\n\n")
	content.Write(Fmt("- 🏆 **Peak Reduction: %.1f%%** (Best optimization)\n", maxImprovement))
	if wasmCount > 0 {
		content.Write(Fmt("- ✅ **Average WebAssembly Reduction: %.1f%%**\n", avgWasmImprovement))
	}
	if nativeCount > 0 {
		content.Write(Fmt("- ✅ **Average Native Reduction: %.1f%%**\n", avgNativeImprovement))
	}
	content.Write(Fmt("- 📦 **Total Size Savings: %s across all builds**\n\n", FormatSize(totalSavings)))

	content.Write("#### Performance Legend\n")
	content.Write("- ❌ Poor (<5% reduction)\n")
	content.Write("- ➖ Fair (5-15% reduction)\n")
	content.Write("- ✅ Good (15-70% reduction)\n")
	content.Write("- 🏆 Outstanding (>70% reduction)\n\n")

	return content.String(), nil
}

// generateMemorySection creates the memory allocation comparison section
func (r *ReportGenerator) generateMemorySection(comparisons []MemoryComparison) (string, error) {
	content := Convert() // Using fmt builder for better performance

	content.Write("## Memory Usage Comparison\n\n")
	content.Write("[Standard Library Example](bench-memory-alloc/standard) | [fmt Example](bench-memory-alloc/tinystring)\n\n")
	content.Write("<!-- This table is automatically generated from memory-benchmark.sh -->\n")
	content.Write("*Last updated: " + time.Now().Format("2006-01-02 15:04:05") + "*\n\n")
	content.Write("Performance benchmarks comparing memory allocation patterns between standard Go library and fmt:\n\n")

	// Enhanced table with better styling and icons
	content.Write("| 🧪 **Benchmark Category** | 📚 **Library** | 💾 **Memory/Op** | 🔢 **Allocs/Op** | ⏱️ **Time/Op** | 📈 **Memory Trend** | 🎯 **Alloc Trend** | 🏆 **Performance** |\n")
	content.Write("|----------------------------|----------------|-------------------|-------------------|-----------------|---------------------|---------------------|--------------------|\n")

	var totalMemoryDiff float64
	var totalAllocDiff float64
	var benchmarkCount int

	for _, comparison := range comparisons {
		if comparison.Standard.Name != "" && comparison.fmt.Name != "" {
			memImprovement := calculateMemoryImprovement(
				comparison.Standard.BytesPerOp, comparison.fmt.BytesPerOp)
			allocImprovement := calculateMemoryImprovement(
				comparison.Standard.AllocsPerOp, comparison.fmt.AllocsPerOp)

			// Calculate percentage changes for tracking
			memPercent := calculateMemoryPercent(comparison.Standard.BytesPerOp, comparison.fmt.BytesPerOp)
			allocPercent := calculateMemoryPercent(comparison.Standard.AllocsPerOp, comparison.fmt.AllocsPerOp)

			totalMemoryDiff += memPercent
			totalAllocDiff += allocPercent
			benchmarkCount++

			// Get performance indicators
			memoryIndicator := getMemoryPerformanceIndicator(memPercent)
			allocIndicator := getAllocPerformanceIndicator(allocPercent)
			overallIndicator := getOverallPerformanceIndicator(memPercent, allocPercent)

			// Category with emoji
			categoryIcon := getBenchmarkCategoryIcon(comparison.Category) // Standard library row with enhanced styling
			// Format operation count with thousands separator
			fmt.Println("DEBUG: Standard Iterations:", comparison.Standard.Iterations)

			opCountStd := Convert(comparison.Standard.Iterations).Thousands().String()
			opCountTiny := Convert(comparison.fmt.Iterations).Thousands().String()

			// Memory/op string with operation count
			memStd := Fmt("%s / %s OP", FormatSize(comparison.Standard.BytesPerOp), opCountStd)
			memTiny := Fmt("%s / %s OP", FormatSize(comparison.fmt.BytesPerOp), opCountTiny)

			content.Write(Fmt("| %s **%s** | 📊 Standard | `%s` | `%d` | `%s` | - | - | - |\n",
				categoryIcon,
				comparison.Category,
				memStd,
				comparison.Standard.AllocsPerOp,
				formatNanoTime(comparison.Standard.NsPerOp)))

			// fmt row with improvements and visual indicators
			content.Write(Fmt("| | 🚀 fmt | `%s` | `%d` | `%s` | %s **%s** | %s **%s** | %s |\n",
				memTiny,
				comparison.fmt.AllocsPerOp,
				formatNanoTime(comparison.fmt.NsPerOp),
				memoryIndicator, memImprovement,
				allocIndicator, allocImprovement,
				overallIndicator))
		}
	}

	// Calculate averages for summary
	var avgMemoryDiff, avgAllocDiff float64
	if benchmarkCount > 0 {
		avgMemoryDiff = totalMemoryDiff / float64(benchmarkCount)
		avgAllocDiff = totalAllocDiff / float64(benchmarkCount)
	}
	// Performance summary section with enhanced styling
	content.Write("\n### 🎯 Performance Summary\n\n")

	// Memory efficiency classification
	memoryClass := getMemoryEfficiencyClass(avgMemoryDiff)
	allocClass := getAllocEfficiencyClass(avgAllocDiff)

	content.Write(Fmt("- 💾 **Memory Efficiency**: %s (%.1f%% average change)\n", memoryClass, avgMemoryDiff))
	content.Write(Fmt("- 🔢 **Allocation Efficiency**: %s (%.1f%% average change)\n", allocClass, avgAllocDiff))
	content.Write(Fmt("- 📊 **Benchmarks Analyzed**: %d categories\n", benchmarkCount))
	content.Write("- 🎯 **Optimization Focus**: Binary size reduction vs runtime efficiency\n\n")

	// Enhanced trade-offs analysis with better formatting
	content.Write("### ⚖️ Trade-offs Analysis\n\n")
	content.Write("The benchmarks reveal important trade-offs between **binary size** and **runtime performance**:\n\n")

	content.Write("#### 📦 **Binary Size Benefits** ✅\n")
	content.Write("- 🏆 **16-84% smaller** compiled binaries\n")
	content.Write("- 🌐 **Superior WebAssembly** compression ratios\n")
	content.Write("- 🚀 **Faster deployment** and distribution\n")
	content.Write("- 💾 **Lower storage** requirements\n\n")

	content.Write("#### 🧠 **Runtime Memory Considerations** ⚠️\n")
	content.Write("- 📈 **Higher allocation overhead** during execution\n")
	content.Write("- 🗑️ **Increased GC pressure** due to allocation patterns\n")
	content.Write("- ⚡ **Trade-off optimizes** for distribution size over runtime efficiency\n")
	content.Write("- 🔄 **Different optimization strategy** than standard library\n\n")
	content.Write("#### 🎯 **Optimization Recommendations**\n")
	content.Write("| 🎯 **Use Case** | 💡 **Recommendation** | 🔧 **Best For** |\n")
	content.Write("|-----------------|------------------------|------------------|\n")
	content.Write("| 🌐 WebAssembly Apps | ✅ **fmt** | Size-critical web deployment |\n")
	content.Write("| 📱 Embedded Systems | ✅ **fmt** | Resource-constrained devices |\n")
	content.Write("| ☁️ Edge Computing | ✅ **fmt** | Fast startup and deployment |\n")
	content.Write("| 🏢 Memory-Intensive Server | ⚠️ **Standard Library** | High-throughput applications |\n")
	content.Write("| 🔄 High-Frequency Processing | ⚠️ **Standard Library** | Performance-critical workloads |\n\n")

	content.Write("#### 📊 **Performance Legend**\n")
	content.Write("- 🏆 **Excellent** (Better performance)\n")
	content.Write("- ✅ **Good** (Acceptable trade-off)\n")
	content.Write("- ⚠️ **Caution** (Higher resource usage)\n")
	content.Write("- ❌ **Poor** (Significant overhead)\n\n")

	return content.String(), nil
}

// updateREADMESection updates a specific section in the README
func (r *ReportGenerator) updateREADMESection(sectionTitle, newContent string) error {
	// Read current README
	existingContent, err := os.ReadFile(r.ReadmePath)
	if err != nil {
		LogError(Fmt("Failed to read README: %v", err))
		return err
	}

	content := string(existingContent)

	// Find section boundaries
	sectionStart := "## " + sectionTitle
	startIndex := strings.Index(content, sectionStart)

	if startIndex == -1 {
		// Section doesn't exist, append to end
		content += "\n" + newContent
	} else {
		// Find next section or end of file
		nextSectionIndex := strings.Index(content[startIndex+len(sectionStart):], "\n## ")
		var endIndex int

		if nextSectionIndex == -1 {
			endIndex = len(content)
		} else {
			endIndex = startIndex + len(sectionStart) + nextSectionIndex
		}

		// Replace the section
		content = content[:startIndex] + newContent + content[endIndex:]
	}

	// Write updated content
	err = os.WriteFile(r.TempPath, []byte(content), 0644)
	if err != nil {
		LogError(Fmt("Failed to write temporary README: %v", err))
		return err
	}

	// Replace original with temporary
	err = os.Rename(r.TempPath, r.ReadmePath)
	if err != nil {
		LogError(Fmt("Failed to replace README: %v", err))
		return err
	}

	LogSuccess(Fmt("Updated README section: %s", sectionTitle))
	return nil
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 || HasUpperPrefix(s) {
		return s
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// Helper functions for binary size reporting

// getBuildTypeIcon returns the appropriate icon for build type
func getBuildTypeIcon(optName string) string {
	switch optName {
	case "Default":
		return "🖥️"
	case "Speed":
		return "⚡"
	case "Ultra":
		return "🏁"
	case "Debug":
		return "🔧"
	default:
		return "📦"
	}
}

// getBuildParameters returns the build parameters for different optimization levels
func getBuildParameters(optName string, isWasm bool) string {
	switch optName {
	case "Default":
		if isWasm {
			return "(default -opt=z)"
		}
		return `-ldflags="-s -w"`
	case "Speed":
		if isWasm {
			return "-opt=2 -target wasm"
		}
		return `-ldflags="-s -w"`
	case "Ultra":
		if isWasm {
			return "-no-debug -panic=trap -scheduler=none -gc=leaking -target wasm"
		}
		return `-ldflags="-s -w"`
	case "Debug":
		if isWasm {
			return "-opt=0 -target wasm"
		}
		return `-ldflags="-s -w"`
	default:
		return ""
	}
}

// calculateImprovementPercent calculates the percentage improvement
func calculateImprovementPercent(standardSize, tinystringSize int64) float64 {
	if standardSize <= 0 {
		return 0
	}
	return float64(standardSize-tinystringSize) / float64(standardSize) * 100
}

// getPerformanceIndicator returns the appropriate performance indicator
func getPerformanceIndicator(improvementPercent float64) string {
	switch {
	case improvementPercent < 5:
		return "❌"
	case improvementPercent < 15:
		return "➖"
	case improvementPercent < 70:
		return "✅"
	default:
		return "🏆"
	}
}

// Helper functions for enhanced memory reporting

// calculateMemoryPercent calculates the percentage change in memory usage
func calculateMemoryPercent(standardValue, tinystringValue int64) float64 {
	if standardValue <= 0 {
		return 0
	}
	return float64(tinystringValue-standardValue) / float64(standardValue) * 100
}

// getBenchmarkCategoryIcon returns appropriate icon for benchmark category
func getBenchmarkCategoryIcon(category string) string {
	switch {
	case strings.Contains(category, "String"):
		return "📝"
	case strings.Contains(category, "Number"):
		return "🔢"
	case strings.Contains(category, "Mixed"):
		return "🔄"
	case strings.Contains(category, "Pointer"):
		return "👉"
	default:
		return "🧪"
	}
}

// getMemoryPerformanceIndicator returns indicator for memory performance
func getMemoryPerformanceIndicator(percentChange float64) string {
	switch {
	case percentChange < -20: // 20% improvement (less memory)
		return "🏆"
	case percentChange < -5: // 5% improvement
		return "✅"
	case percentChange < 5: // Similar usage
		return "➖"
	case percentChange < 50: // ToUpper to 50% more
		return "⚠️"
	default: // Over 50% more
		return "❌"
	}
}

// getAllocPerformanceIndicator returns indicator for allocation performance
func getAllocPerformanceIndicator(percentChange float64) string {
	switch {
	case percentChange < -15: // 15% fewer allocations
		return "🏆"
	case percentChange < -5: // 5% fewer allocations
		return "✅"
	case percentChange < 5: // Similar allocations
		return "➖"
	case percentChange < 25: // ToUpper to 25% more
		return "⚠️"
	default: // Over 25% more
		return "❌"
	}
}

// getOverallPerformanceIndicator combines memory and allocation indicators
func getOverallPerformanceIndicator(memPercent, allocPercent float64) string {
	// Average the two percentages for overall assessment
	avgChange := (memPercent + allocPercent) / 2

	switch {
	case avgChange < -15: // Overall improvement
		return "🏆 **Excellent**"
	case avgChange < -5: // Slight improvement
		return "✅ **Good**"
	case avgChange < 15: // Acceptable trade-off
		return "➖ **Fair**"
	case avgChange < 40: // Higher resource usage
		return "⚠️ **Caution**"
	default: // Significant overhead
		return "❌ **Poor**"
	}
}

// getMemoryEfficiencyClass classifies memory efficiency
func getMemoryEfficiencyClass(avgPercent float64) string {
	switch {
	case avgPercent < -10:
		return "🏆 **Excellent** (Lower memory usage)"
	case avgPercent < 0:
		return "✅ **Good** (Memory efficient)"
	case avgPercent < 20:
		return "➖ **Fair** (Acceptable overhead)"
	case avgPercent < 50:
		return "⚠️ **Caution** (Higher memory usage)"
	default:
		return "❌ **Poor** (Significant overhead)"
	}
}

// getAllocEfficiencyClass classifies allocation efficiency
func getAllocEfficiencyClass(avgPercent float64) string {
	switch {
	case avgPercent < -10:
		return "🏆 **Excellent** (Fewer allocations)"
	case avgPercent < 0:
		return "✅ **Good** (Allocation efficient)"
	case avgPercent < 15:
		return "➖ **Fair** (Acceptable allocation pattern)"
	case avgPercent < 35:
		return "⚠️ **Caution** (More allocations)"
	default:
		return "❌ **Poor** (Excessive allocations)"
	}
}

// calculateMemoryImprovement formats the memory improvement as a string
func calculateMemoryImprovement(standardValue, tinystringValue int64) string {
	if standardValue <= 0 {
		return "N/A"
	}

	percent := float64(standardValue-tinystringValue) / float64(standardValue) * 100
	if percent > 0 {
		return Fmt("%.1f%% less", percent)
	} else if percent < 0 {
		return Fmt("%.1f%% more", -percent)
	}
	return "Same"
}

// formatNanoTime formats nanoseconds to readable time units
func formatNanoTime(ns int64) string {
	if ns < 1000 {
		return Fmt("%dns", ns)
	} else if ns < 1000000 {
		return Fmt("%.1fμs", float64(ns)/1000)
	} else {
		return Fmt("%.1fms", float64(ns)/1000000)
	}
}
