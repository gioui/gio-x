package org.gioui.x.share;

import android.app.Activity;
import android.view.View;
import android.content.Intent;

public class share_android {
    static public void openShare(View view, String title, Intent i) {
        ((Activity) view.getContext()).startActivity(Intent.createChooser(i, title));
    }
    static public void shareText(View view, String title, String text) {
        Intent i = new Intent(Intent.ACTION_SEND);
        i.setType("text/plain");
        i.putExtra(Intent.EXTRA_TEXT, text);

        openShare(view, title, i);
    }
    static public void shareWebsite(View view, String title, String text, String link) {
        Intent i = new Intent(Intent.ACTION_SEND);
        i.setType("text/plain");
        i.putExtra(Intent.EXTRA_SUBJECT, text);
        i.putExtra(Intent.EXTRA_TEXT, link);

        openShare(view, title, i);
    }
}