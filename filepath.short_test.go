package fmt

import (
	"os"
	"testing"
)

func TestPathShort(t *testing.T) {
	// Setup base for tests
	originalBase := pathBase
	defer func() { pathBase = originalBase }()

	wd, _ := os.Getwd()

	tests := []struct {
		name string
		base string
		path string
		want string
	}{
		{
			name: "relative from explicit wd",
			base: wd, // Use explicit wd instead of auto-detection (WASM auto-detects URL origin)
			path: PathJoin(wd, "web/public").String(),
			want: "./web/public",
		},
		{
			name: "relative from manual base",
			base: "/home/user/project",
			path: "/home/user/project/modules/test.js",
			want: "./modules/test.js",
		},
		{
			name: "exactly same as base",
			base: "/home/user/project",
			path: "/home/user/project",
			want: ".",
		},
		{
			name: "different path",
			base: "/home/user/project",
			path: "/etc/passwd",
			want: "/etc/passwd",
		},
		{
			name: "prefix but not subpath",
			base: "/home/user/pro",
			path: "/home/user/project",
			want: "/home/user/project",
		},
		{
			name: "subpath with trailing slash in input",
			base: "/home/user/project",
			path: "/home/user/project/web/",
			want: "./web/",
		},
		{
			name: "manually set base as root",
			base: "/",
			path: "/etc/passwd",
			want: "./etc/passwd",
		},
		{
			name: "embedded path in log message",
			base: "/home/user/Dev/Pkg/tinywasm/app/example",
			path: "Compiling WASM due to /home/user/Dev/Pkg/tinywasm/app/example/web/client.go change... ",
			want: "Compiling WASM due to ./web/client.go change... ",
		},
		{
			name: "another embedded path in log message",
			base: "/home/user/Dev/Pkg/tinywasm/app/example",
			path: " 13:07:52  ASSETS  .js create ... /home/user/Dev/Pkg/tinywasm/app/example/modules/users/newfile.js",
			want: " 13:07:52  ASSETS  .js create ... ./modules/users/newfile.js",
		},
		{
			name: "path at the end of sentence",
			base: "/home/user/Dev/Pkg/tinywasm/app/example",
			path: "WASM source file already exists at /home/user/Dev/Pkg/tinywasm/app/example/web/client.go , skipping generation",
			want: "WASM source file already exists at ./web/client.go , skipping generation",
		},
		{
			name: "multiple occurrences",
			base: "/home/user/project",
			path: "moving /home/user/project/a to /home/user/project/b",
			want: "moving ./a to ./b",
		},
		{
			name: "within quotes",
			base: "/home/user/project",
			path: `source is "/home/user/project/main.go"`,
			want: `source is "./main.go"`,
		},
		{
			name: "base followed by punctuation",
			base: "/home/user/project",
			path: "current dir is /home/user/project, check it.",
			want: "current dir is /home/user/project, check it.", // NOT valid boundary because , is not / or \
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.base != "" {
				SetPathBase(tc.base)
			} else {
				pathBase = "" // trigger auto-detection
			}

			// PathShort expects the path to be in BuffOut
			got := Convert(tc.path).PathShort().String()

			// adjust expected for auto-detection case if needed
			want := tc.want
			if tc.base == "" {
				// if we used auto-detection, the want is relative to cleanWD
				// our test setup uses PathJoin(wd, "web/public") so it should match
			}

			if got != want {
				t.Errorf("%s: PathShort(%q) with base %q = %q; want %q", tc.name, tc.path, pathBase, got, want)
			}
		})
	}
}

func TestPathShortWindows(t *testing.T) {
	// Manual test for windows-style paths even on linux
	// since pathClean and PathJoin handle them conceptually
	originalBase := pathBase
	defer func() { pathBase = originalBase }()

	SetPathBase(`C:\Users\Project`)

	got := Convert(`C:\Users\Project\file.txt`).PathShort().String()
	want := "./file.txt"
	if got != want {
		t.Errorf("Windows relative: got %q; want %q", got, want)
	}

	got = Convert(`C:\Users\Project`).PathShort().String()
	want = "."
	if got != want {
		t.Errorf("Windows same: got %q; want %q", got, want)
	}
}
