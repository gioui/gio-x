package battery

import (
	"gioui.org/app"
	"gioui.org/x/pref/internal/xjni"
	"git.wow.st/gmp/jni"
)

//go:generate javac -source 8 -target 8 -bootclasspath $ANDROID_HOME/platforms/android-29/android.jar -d $TEMP/pref_battery/classes battery_android.java
//go:generate jar cf battery_android.jar -C $TEMP/pref_battery/classes .

var _Lib = "org/gioui/x/pref/battery/battery_android"

func batteryLevel() (uint8, error) {
	i, err := xjni.DoInt(_Lib, "batteryLevel", "(Landroid/content/Context;)I", jni.Value(app.AppContext()))
	if err != nil || i < 0 {
		return 100, ErrNotAvailableAPI
	}
	return uint8(i), nil
}

func isSavingBattery() (bool, error) {
	i, err := xjni.DoInt(_Lib, "isSaving", "(Landroid/content/Context;)I", jni.Value(app.AppContext()))
	if err != nil || i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, err
}

func isCharging() (bool, error) {
	i, err := xjni.DoInt(_Lib, "isCharging", "(Landroid/content/Context;)I", jni.Value(app.AppContext()))
	if err != nil || i < 0 {
		return false, ErrNotAvailableAPI
	}
	return i >= 1, err
}
