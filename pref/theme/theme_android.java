package org.gioui.x.pref.theme;

import android.content.res.Resources;
import android.content.res.Configuration;
import android.provider.Settings.Global;
import android.os.Build;

public class theme_android {
	public static int isDark() {
        Resources res = Resources.getSystem();
        if (res == null) {
            return -1;
        }

        Configuration config = res.getConfiguration();
        if (config == null) {
            return -1;
        }

        if ((config.uiMode & Configuration.UI_MODE_NIGHT_MASK) == Configuration.UI_MODE_NIGHT_YES) {
           return 1;
        }

        return 0;
	}

	public static int isReducedMotion() {
	    if (Global.TRANSITION_ANIMATION_SCALE == "0") {
	        return 1;
	    }
	    return 0;
	}
}