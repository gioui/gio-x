package ht.sr.git.whereswaldon.niotify;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.util.Log;
import android.graphics.Bitmap;
import android.graphics.drawable.Icon;

public class NotificationHelper {
    public static void newChannel() {
        Log.w("NotificationHelper",String.format("newChannel invoked"));
    }
}
