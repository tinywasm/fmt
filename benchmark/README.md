# fmt Benchmark Suite

Automated benchmark tools to measure and compare performance between standard Go libraries and fmt implementations.



## Binary Size Comparison

[Standard Library Example](bench-binary-size/standard-lib/main.go) | [fmt Example](bench-binary-size/tinystring-lib/main.go)

<!-- This table is automatically generated from build-and-measure.sh -->
*Last updated: 2026-02-11 07:51:43*

| Build Type | Parameters | Standard Library<br/>`go build` | fmt<br/>`tinygo build` | Size Reduction | Performance |
|------------|------------|------------------|------------|----------------|-------------|
| 🖥️ **Default Native** | `-ldflags="-s -w"` | 1.4 MB | 1.3 MB | **-132.0 KB** | ➖ **8.9%** |
| 🌐 **Default WASM** | `(default -opt=z)` | 606.9 KB | 271.1 KB | **-335.9 KB** | ✅ **55.3%** |
| 🌐 **Ultra WASM** | `-no-debug -panic=trap -scheduler=none -gc=leaking -target wasm` | 155.6 KB | 30.4 KB | **-125.1 KB** | 🏆 **80.4%** |
| 🌐 **Speed WASM** | `-opt=2 -target wasm` | 868.5 KB | 441.6 KB | **-426.9 KB** | ✅ **49.2%** |
| 🌐 **Debug WASM** | `-opt=0 -target wasm` | 1.8 MB | 872.3 KB | **-973.5 KB** | ✅ **52.7%** |

### 🎯 Performance Summary

- 🏆 **Peak Reduction: 80.4%** (Best optimization)
- ✅ **Average WebAssembly Reduction: 59.4%**
- ✅ **Average Native Reduction: 8.9%**
- 📦 **Total Size Savings: 1.9 MB across all builds**

#### Performance Legend
- ❌ Poor (<5% reduction)
- ➖ Fair (5-15% reduction)
- ✅ Good (15-70% reduction)
- 🏆 Outstanding (>70% reduction)


## Memory Usage Comparison

[Standard Library Example](bench-memory-alloc/standard) | [fmt Example](bench-memory-alloc/tinystring)

<!-- This table is automatically generated from memory-benchmark.sh -->
*Last updated: 2026-02-11 07:51:58*

Performance benchmarks comparing memory allocation patterns between standard Go library and fmt:

| 🧪 **Benchmark Category** | 📚 **Library** | 💾 **Memory/Op** | 🔢 **Allocs/Op** | ⏱️ **Time/Op** | 📈 **Memory Trend** | 🎯 **Alloc Trend** | 🏆 **Performance** |
|----------------------------|----------------|-------------------|-------------------|-----------------|---------------------|---------------------|--------------------|
| 📝 **String Processing** | 📊 Standard | `808 B / 523.990 OP` | `32` | `2.1μs` | - | - | - |
| | 🚀 fmt | `464 B / 227.250 OP` | `17` | `5.5μs` | 🏆 **42.6% less** | 🏆 **46.9% less** | 🏆 **Excellent** |
| 🔢 **Number Processing** | 📊 Standard | `720 B / 527.766 OP` | `34` | `2.3μs` | - | - | - |
| | 🚀 fmt | `320 B / 523.962 OP` | `17` | `2.2μs` | 🏆 **55.6% less** | 🏆 **50.0% less** | 🏆 **Excellent** |
| 🔄 **Mixed Operations** | 📊 Standard | `368 B / 772.447 OP` | `20` | `1.5μs` | - | - | - |
| | 🚀 fmt | `192 B / 469.515 OP` | `12` | `2.6μs` | 🏆 **47.8% less** | 🏆 **40.0% less** | 🏆 **Excellent** |

### 🎯 Performance Summary

- 💾 **Memory Efficiency**: 🏆 **Excellent** (Lower memory usage) (-48.7% average change)
- 🔢 **Allocation Efficiency**: 🏆 **Excellent** (Fewer allocations) (-45.6% average change)
- 📊 **Benchmarks Analyzed**: 3 categories
- 🎯 **Optimization Focus**: Binary size reduction vs runtime efficiency

### ⚖️ Trade-offs Analysis

The benchmarks reveal important trade-offs between **binary size** and **runtime performance**:

#### 📦 **Binary Size Benefits** ✅
- 🏆 **16-84% smaller** compiled binaries
- 🌐 **Superior WebAssembly** compression ratios
- 🚀 **Faster deployment** and distribution
- 💾 **Lower storage** requirements

#### 🧠 **Runtime Memory Considerations** ⚠️
- 📈 **Higher allocation overhead** during execution
- 🗑️ **Increased GC pressure** due to allocation patterns
- ⚡ **Trade-off optimizes** for distribution size over runtime efficiency
- 🔄 **Different optimization strategy** than standard library

#### 🎯 **Optimization Recommendations**
| 🎯 **Use Case** | 💡 **Recommendation** | 🔧 **Best For** |
|-----------------|------------------------|------------------|
| 🌐 WebAssembly Apps | ✅ **fmt** | Size-critical web deployment |
| 📱 Embedded Systems | ✅ **fmt** | Resource-constrained devices |
| ☁️ Edge Computing | ✅ **fmt** | Fast startup and deployment |
| 🏢 Memory-Intensive Server | ⚠️ **Standard Library** | High-throughput applications |
| 🔄 High-Frequency Processing | ⚠️ **Standard Library** | Performance-critical workloads |

#### 📊 **Performance Legend**
- 🏆 **Excellent** (Better performance)
- ✅ **Good** (Acceptable trade-off)
- ⚠️ **Caution** (Higher resource usage)
- ❌ **Poor** (Significant overhead)


## Quick Usage 🚀

```bash
# Run complete benchmark (recommended)
./build-and-measure.sh

# Clean generated files
./clean-all.sh

# Update README with existing data only (does not re-run benchmarks)
./update-readme.sh

# Run all memory and binary size benchmarks (without updating README)
./run-all-benchmarks.sh

# Run only memory benchmarks
./memory-benchmark.sh
```

## What Gets Measured 📊

1.  **Binary Size Comparison**: Native + WebAssembly builds with multiple optimization levels. This compares the compiled output size of projects using the standard Go library versus fmt.
2.  **Memory Allocation**: Measures Bytes/op, Allocations/op, and execution time (ns/op) for benchmark categories. This helps in understanding the memory efficiency of fmt compared to standard library operations.
    *   **String Processing**: Benchmarks operations like case conversion, text manipulation, etc.
    *   **Number Processing**: Benchmarks numeric formatting, conversion operations, etc.
    *   **Mixed Operations**: Benchmarks scenarios involving a combination of string and numeric operations.

## Current Performance Status

**Target**: Achieve memory usage close to standard library while maintaining binary size benefits.

**Latest Results** (Run `./build-and-measure.sh` to update):
- ✅ **Binary Size**: fmt is 20-50% smaller than stdlib for WebAssembly.
- ⚠️ **Memory Usage**: Number Processing uses 1000% more memory (needs optimization).

📋 **Memory Optimization Guide**: See [`MEMORY_REDUCTION.md`](./MEMORY_REDUCTION.md) for comprehensive techniques and best practices to replace Go standard libraries with fmt's optimized implementations. Essential reading for efficient string and numeric processing in TinyGo WebAssembly applications.

## Requirements

- **Go 1.21+**
- **TinyGo** (optional, but recommended for full WebAssembly testing and to achieve smallest binary sizes).


## Troubleshooting

**TinyGo Not Found:**
```
❌ TinyGo is not installed. Building only standard Go binaries.
```
Install TinyGo from: https://tinygo.org/getting-started/install/

**Permission Issues (Linux/macOS/WSL):**
If you encounter permission errors when trying to run the shell scripts, make them executable:
```bash
chmod +x *.sh
```

**Build Failures:**
- Ensure you're in the `benchmark/` directory
- Verify fmt library is available in the parent directory




