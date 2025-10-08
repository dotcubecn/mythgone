//go:build windows

package main

import (
	"log"
	"syscall"

	"github.com/tailscale/walk"
	. "github.com/tailscale/walk/declarative"
)

// 创建反控窗口
func CreateControlWindow() error {
	if controlWindow != nil {
		controlWindow.Show()
		controlWindow.BringToTop()
		return nil
	}

	err := Dialog{
		AssignTo: &controlWindow,
		Title:    "Mythgone 极域反控",
		Icon: func() *walk.Icon {
			icon, err := walk.NewIconFromResourceId(2)
			if err != nil {
				log.Fatal(err)
			}
			return icon
		}(),
		MinSize:       Size{Width: 200, Height: 200},
		Size:          Size{Width: 200, Height: 200},
		FixedSize:     true,
		Layout:        VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
		CancelButton:  nil,
		DefaultButton: nil,
		Children: []Widget{
			VSpacer{},
			// 底部链接
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					LinkLabel{
						Text: `由 dotcubecn 与所有贡献者开发 (<a href="https://github.com/dotcubecn/mythgone">GitHub</a>)`,
						OnLinkActivated: func(link *walk.LinkLabelLink) {
							OpenURL(link.URL())
						},
					},
					HSpacer{},
				},
			},
		},
	}.Create(mainWindow)

	if err != nil {
		return err
	}
	// 同步主窗口状态
	if preventCaptureCheckbox != nil && preventCaptureCheckbox.Checked() {
		SetWindowDisplayAffinity(syscall.Handle(controlWindow.Handle()), WindowDisplayAffinityExcludeFromCapture)
	}
	if topmostCheckbox != nil && topmostCheckbox.Checked() {
		SetWindowTopmost(syscall.Handle(controlWindow.Handle()), true)
	}
	// 窗口关闭事件处理
	controlWindow.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		controlWindow = nil
	})
	controlWindow.Show()

	return nil
}

// 更新禁止捕获状态
func UpdateControlWindowCaptureState(preventCapture bool) {
	if controlWindow != nil {
		if preventCapture {
			SetWindowDisplayAffinity(syscall.Handle(controlWindow.Handle()), WindowDisplayAffinityExcludeFromCapture)
		} else {
			SetWindowDisplayAffinity(syscall.Handle(controlWindow.Handle()), WindowDisplayAffinityNone)
		}
	}
}
