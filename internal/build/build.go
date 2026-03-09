package build

// Version is injected at build time via -ldflags.
// Default value is used when running without a build step (e.g. go run).
var Version = "v0.0.0"
