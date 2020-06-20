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
    public static void newChannel(Context ctx, int importance, String channelID, String name, String description) {
        NotificationChannel channel = new NotificationChannel(channelID, name, importance);
        channel.setDescription(description);

        NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
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
    public static void cancelNotification(Context ctx, int notificationID) {
        NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
        notificationManager.cancel(notificationID);
    }
}
