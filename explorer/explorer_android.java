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
import java.io.Closeable;
import java.io.Flushable;
import java.util.ArrayList;
import java.util.List;

public class explorer_android {
    final Fragment frag = new explorer_android_fragment();

    static List<Integer> import_codes = new ArrayList<Integer>();
    static List<Integer> export_codes = new ArrayList<Integer>();

    static int PREFIX_CODE = 756456;

    static public native void ImportCallback(InputStream f, int id);
    static public native void ExportCallback(OutputStream f, int id);

    public static class explorer_android_fragment extends Fragment {
        Context context;

        @Override public void onAttach(Context ctx) {
            context = ctx;
            super.onAttach(ctx);
        }

        @Override public void onActivityResult(int requestCode, int resultCode, Intent data) {
            super.onActivityResult(requestCode, resultCode, data);

            Activity activity = this.getActivity();

            activity.runOnUiThread(new Runnable() {
                public void run() {
                    if (import_codes.contains(Integer.valueOf(requestCode))) {
                        import_codes.remove(Integer.valueOf(requestCode));
                        if (resultCode != Activity.RESULT_OK) {
                            explorer_android.ImportCallback(null, requestCode - PREFIX_CODE);
                            activity.getFragmentManager().popBackStack();
                            return;
                        }
                        try {
                            InputStream f = activity.getApplicationContext().getContentResolver().openInputStream(data.getData());
                            explorer_android.ImportCallback(f, requestCode - PREFIX_CODE);
                        } catch (Exception e) {
                            explorer_android.ImportCallback(null, requestCode - PREFIX_CODE);
                            e.printStackTrace();
                            return;
                        }
                    }

                    if (export_codes.contains(Integer.valueOf(requestCode))) {
                        export_codes.remove(Integer.valueOf(requestCode));
                        if (resultCode != Activity.RESULT_OK) {
                            explorer_android.ExportCallback(null, requestCode - PREFIX_CODE);
                            activity.getFragmentManager().popBackStack();
                            return;
                        }
                        try {
                            OutputStream f = activity.getApplicationContext().getContentResolver().openOutputStream(data.getData());
                            explorer_android.ExportCallback(f, requestCode - PREFIX_CODE);
                        } catch (Exception e) {
                            explorer_android.ExportCallback(null, requestCode - PREFIX_CODE);
                            e.printStackTrace();
                            return;
                        }
                    }
                }
            });


        }
    }

    public int fileRead(InputStream f, byte[] b) {
        try {
            return f.read(b, 0, b.length);
        } catch (Exception e) {
            e.printStackTrace();
            return -1;
        }
    }

    public boolean fileWrite(OutputStream f, byte[] b) {
        try {
            f.write(b);
            return true;
        } catch (Exception e) {
            e.printStackTrace();
            return false;
        }
    }

    public boolean fileClose(Closeable c, Flushable f) {
        try {
            if (f != null) {
                f.flush();
            }
            c.close();
            return true;
        } catch (Exception e) {
            e.printStackTrace();
            return false;
        }
    }

    public void exportFile(View view, String ext, int id) {
        askPermission(view);

        ((Activity) view.getContext()).runOnUiThread(new Runnable() {
            public void run() {
                registerFrag(view);
                export_codes.add(Integer.valueOf(id + PREFIX_CODE));
                
                final Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                intent.setType(MimeTypeMap.getSingleton().getMimeTypeFromExtension(ext));
                intent.addCategory(Intent.CATEGORY_OPENABLE);
                frag.startActivityForResult(Intent.createChooser(intent, ""), id + PREFIX_CODE);
            }
        });
    }

    public void importFile(View view, String mime, int id) {
        askPermission(view);

        ((Activity) view.getContext()).runOnUiThread(new Runnable() {
            public void run() {
                registerFrag(view);
                import_codes.add(Integer.valueOf(id + PREFIX_CODE));

                final Intent intent = new Intent(Intent.ACTION_GET_CONTENT);
                intent.setType("*/*");
                intent.addCategory(Intent.CATEGORY_OPENABLE);

                if (mime != null) {
                    final String[] mimes = mime.split(",");
                    if (mimes != null && mimes.length > 0) {
                        intent.putExtra(Intent.EXTRA_MIME_TYPES, mimes);
                    }
                }
                frag.startActivityForResult(Intent.createChooser(intent, ""), id + PREFIX_CODE);
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

        if (fm.findFragmentByTag("explorer_android_fragment") != null) {
            return; // Already exists;
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