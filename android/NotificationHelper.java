package ht.sr.git.whereswaldon.niotify;

import android.content.Context;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.util.Log;
import android.graphics.Bitmap;
import android.graphics.drawable.Icon;

public class NotificationHelper {
    public static NotificationChannel newChannel(Context ctx) {
        String tag = "NotificationHelper";
        Log.w(tag,String.format("newChannel invoked"));
        String CHANNEL_ID = "CHANNEL_ID";
        CharSequence name = "notification_channel_name";
        String description = "notification_channel_description";
        int importance = NotificationManager.IMPORTANCE_DEFAULT;
        NotificationChannel channel = new NotificationChannel(CHANNEL_ID, name, importance);
    	Log.e(tag,String.format("channel: %s",channel));
        channel.setDescription(description);

        NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
    	Log.e(tag,String.format("manager: %s",notificationManager));
        notificationManager.createNotificationChannel(channel);
        return channel;
    }
}
