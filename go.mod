module github.com/dotcubecn/mythgone

go 1.25.1

require github.com/tailscale/walk v0.0.0-20250702155327-6376defdac3f

require (
	github.com/tailscale/win v0.0.0-20250213223159-5992cb43ca35
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/sys v0.36.0
)

require (
	github.com/dblohm7/wingoes v0.0.0-20231019175336-f6e33aa7cc34 // indirect
	golang.org/x/exp v0.0.0-20230425010034-47ecfdc1ba53 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
)

replace github.com/tailscale/walk v0.0.0-20250702155327-6376defdac3f => github.com/dotcubecn/tailscale-walk-toolbar-fix v0.0.0-20251005105426-ed4f3f0af2f9
