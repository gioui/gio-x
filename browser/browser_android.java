package org.gioui.x.browser;

import android.content.Context;
import android.content.Intent;
import android.net.Uri;

public class browser_android {
  public static void openUrl(Context context, String url) {
    Uri webpage = Uri.parse(url);
    Intent intent = new Intent(Intent.ACTION_VIEW, webpage);
    intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
    context.startActivity(intent);
  }
}