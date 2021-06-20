package org.gioui.x.explorer;

import android.content.Context;
import android.util.Log;
import android.content.Intent;
import android.view.View;
import android.app.Activity;
import android.Manifest;
import android.content.pm.PackageManager;
import android.os.Handler.Callback;
import android.os.Handler;
import android.net.Uri;
import android.app.Fragment;
import android.app.FragmentManager;
import android.app.FragmentTransaction;
import android.os.Looper;
import android.content.ContentResolver;
import java.io.InputStream;
import java.io.OutputStream;
import android.webkit.MimeTypeMap;
import java.io.ByteArrayOutputStream;

public class explorer_android {
  final Fragment frag = new explorer_android_fragment();

  static int CODE_OPEN = 756456;
  static int CODE_CREATE = 756457;

  static public native void OpenCallback(byte[] b);
  static public native void CreateCallback(boolean b);

  public static class explorer_android_fragment extends Fragment {
    static byte[] content;

    @Override public void onAttach(Context ctx) {
      super.onAttach(ctx);
    }

    @Override public void onActivityResult(int requestCode, int resultCode, Intent data) {
      super.onActivityResult(requestCode, resultCode, data);

      if (resultCode != Activity.RESULT_OK && (requestCode == explorer_android.CODE_OPEN || requestCode == explorer_android.CODE_CREATE)) {
        if (requestCode == explorer_android.CODE_OPEN) {
          explorer_android.OpenCallback(null);
        }
        if (requestCode == explorer_android.CODE_CREATE) {
          explorer_android.CreateCallback(false);
        }
        return;
      }

      if (requestCode == explorer_android.CODE_OPEN) {
        try {
          InputStream f = this.getActivity().getApplicationContext().getContentResolver().openInputStream(data.getData());
          if (f == null) {
            explorer_android.OpenCallback(null);
          }

          ByteArrayOutputStream buffer = new ByteArrayOutputStream();

          int i;
          byte[] d = new byte[1024];

          while ((i = f.read(d, 0, d.length)) != -1) {
            buffer.write(d, 0, i);
          }

          buffer.flush();

          explorer_android.OpenCallback(buffer.toByteArray());
        } catch (Exception e) {
          e.printStackTrace();
          return;
        }
      }

      if (requestCode == explorer_android.CODE_CREATE) {
        try {
          OutputStream f = this.getActivity().getApplicationContext().getContentResolver().openOutputStream(data.getData());
          if (f == null) {
            explorer_android.CreateCallback(false);
          }

          f.write(this.content);
          f.flush();
          f.close();

          explorer_android.CreateCallback(true);
        } catch (Exception e) {
          e.printStackTrace();
          return;
        }
      }

    }
  }

  public void openFile(View view, String mime) {
    askPermission(view);

    new Handler(Looper.getMainLooper()).post(new Runnable() {
      public void run() {
        registerFrag(view);

        final Intent intent = new Intent(Intent.ACTION_GET_CONTENT);
        intent.setType("*/*");
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        if (mime != null && mime.split(",").length > 0) {
          intent.putExtra(Intent.EXTRA_MIME_TYPES, mime.split(","));
        }
        frag.startActivityForResult(Intent.createChooser(intent, ""), explorer_android.CODE_OPEN);
      }
    });
  }

  public void createFile(View view, String ext, byte[] content) {
    askPermission(view);

    new Handler(Looper.getMainLooper()).post(new Runnable() {
      public void run() {
        registerFrag(view);

        final Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
        intent.setType(MimeTypeMap.getSingleton().getMimeTypeFromExtension(ext));
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        explorer_android_fragment.content = content;
        frag.startActivityForResult(Intent.createChooser(intent, ""), explorer_android.CODE_CREATE);
      }
    });
  }

  public void registerFrag(View view) {
    final Context ctx = view.getContext();
    final FragmentManager fm;

    try {
      fm = (FragmentManager) ctx.getClass().getMethod("getFragmentManager").invoke(ctx);
    } catch (Exception e) {
      e.printStackTrace();
      return;
    }

    FragmentTransaction ft = fm.beginTransaction();
    ft.add(frag, "explorer_android_fragment");
    ft.commitNow();
  }

  public void askPermission(View view) {
    Activity activity = (Activity) view.getContext();

    if (activity.checkSelfPermission(Manifest.permission.READ_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
      activity.requestPermissions(new String[] {
        Manifest.permission.READ_EXTERNAL_STORAGE
      }, 255);
    }

    if (activity.checkSelfPermission(Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
      activity.requestPermissions(new String[] {
        Manifest.permission.WRITE_EXTERNAL_STORAGE
      }, 254);
    }
  }
}