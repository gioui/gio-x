package ht.sr.git.whereswaldon.niotify;

import android.content.Context;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.util.Log;
import android.graphics.Bitmap;
import android.graphics.drawable.Icon;

public class NotificationHelper {
    private final static String tag = "NotificationHelper";
    public static void newChannel(Context ctx, String channelID, String name, String description) {
        Log.w(tag,String.format("newChannel invoked"));
        int importance = NotificationManager.IMPORTANCE_DEFAULT;
        NotificationChannel channel = new NotificationChannel(channelID, name, importance);
    	Log.e(tag,String.format("channel: %s",channel));
        channel.setDescription(description);

        NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
    	Log.e(tag,String.format("manager: %s",notificationManager));
        notificationManager.createNotificationChannel(channel);
    }
    public static void sendNotification(Context ctx, String channelID, int notificationID, String title, String text) {
        Notification.Builder builder = new Notification.Builder(ctx, channelID)
                .setContentTitle(title)
                .setSmallIcon(Icon.createWithBitmap(Bitmap.createBitmap(1,1,Bitmap.Config.ALPHA_8)))
                .setContentText(text)
                .setPriority(Notification.PRIORITY_DEFAULT);

        NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
        notificationManager.notify(notificationID, builder.build());
    }
}
