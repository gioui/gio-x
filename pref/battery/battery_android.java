package org.gioui.x.pref.battery;

import android.os.Build;
import android.content.Context;
import android.os.BatteryManager;
import android.os.PowerManager;
import android.util.Log;

public class battery_android {
    public static int batteryLevel(Context context) {
        if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.LOLLIPOP) {
            return -1;
        }

        BatteryManager bm = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);
        if (bm == null) {
            return -1;
        }

        return bm.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY);
    }

    public static int isCharging(Context context) {
        if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.LOLLIPOP) {
            return -1;
        }

        BatteryManager bm = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);
        if (bm == null) {
            return -1;
        }

        if (bm.isCharging()) {
            return 1;
        }
        return 0;
    }

    public static int isSaving(Context context) {
        if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.LOLLIPOP) {
            return -1;
        }

        PowerManager pm = (PowerManager) context.getSystemService(Context.POWER_SERVICE);
        if (pm == null) {
            return -1;
        }

        if (pm.isPowerSaveMode()) {
            return 1;
        }
        return 0;
    }
}