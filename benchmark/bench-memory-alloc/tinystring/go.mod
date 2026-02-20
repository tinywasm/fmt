module memory-bench-tinystring

go 1.25.2

require (
	benchmark/shared v0.0.0
	github.com/tinywasm/fmt v0.18.0
)

// Use local fmt module

// Use local shared module
replace benchmark/shared => ../../shared
