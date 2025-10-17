module github.com/dotcubecn/mythgone

go 1.25.1

require github.com/tailscale/walk v0.0.0-20251016200523-963e260a8227

require (
	github.com/tailscale/win v0.0.0-20250627215312-f4da2b8ee071
	golang.org/x/net v0.46.0
	golang.org/x/sys v0.37.0
)

require (
	github.com/dblohm7/wingoes v0.0.0-20250822163801-6d8e6105c62d // indirect
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
)

replace github.com/tailscale/walk v0.0.0-20251016200523-963e260a8227 => github.com/dotcubecn/tailscale-walk-toolbar-fix v0.0.0-20251017110751-499fd12bf88e
