//go:build !linux

package gui

// webkitDefaults: only Linux's webview is WebKitGTK.
func webkitDefaults() {}

// webkitPlaysVideo: the other systems' webviews play video.
func webkitPlaysVideo() bool { return true }
