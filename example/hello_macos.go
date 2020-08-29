//go:generate go build
//go:generate cp -f example example.app/Contents/MacOS
//go:generate codesign -s - example.app

package main
