package org.gioui.x.pref.locale;

import android.content.res.Resources;
import android.os.Build;
import java.util.Locale;

public class locale_android {
	public static String getLanguage() {
	    if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.LOLLIPOP) {
	        return "";
	    }

        if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.N) {
           return Resources.getSystem().getConfiguration().locale.toLanguageTag();
        }

        return Resources.getSystem().getConfiguration().getLocales().get(0).toLanguageTag();
	}
}