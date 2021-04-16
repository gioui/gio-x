package locale

import (
	"gioui.org/x/pref/internal/xjni"
)

//go:generate javac -source 8 -target 8 -bootclasspath $ANDROID_HOME/platforms/android-29/android.jar -d $TEMP/x_locale/classes locale_android.java
//go:generate jar cf locale_android.jar -C $TEMP/x_locale/classes .

var (
	_Lib = "org/gioui/x/pref/locale/locale_android"
)

func getLanguage() string {
	lang, err := xjni.DoString(_Lib, "getLanguage", "()Ljava/lang/String;")
	if err != nil {
		return ""
	}

	return lang
}
