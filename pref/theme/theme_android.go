package theme

//go:generate javac -source 8 -target 8 -bootclasspath $ANDROID_HOME/platforms/android-29/android.jar -d $TEMP/pref_theme/classes theme_android.java
//go:generate jar cf theme_android.jar -C $TEMP/pref_theme/classes .

import (
	"gioui.org/x/pref/internal/xjni"
)

var _Lib = "org/gioui/x/pref/theme/theme_android"

func isDark() (bool, error) {
	i, err := xjni.DoInt(_Lib, "isDark", "()I")
	if err != nil || i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, nil
}

func isReducedMotion() (bool, error) {
	i, err := xjni.DoInt(_Lib, "isReducedMotion", "()I")
	if err != nil || i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, nil
}
