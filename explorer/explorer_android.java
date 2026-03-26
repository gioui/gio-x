package org.gioui.x.explorer;

import android.content.ContentResolver;
import android.os.ParcelFileDescriptor;
import android.util.Log;
import android.content.Context;
import android.content.Intent;
import android.database.Cursor;
import android.provider.OpenableColumns;
import android.view.View;
import android.app.Activity;
import android.Manifest;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.app.Fragment;
import android.app.FragmentManager;
import android.app.FragmentTransaction;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;

import android.webkit.MimeTypeMap;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

@SuppressWarnings("deprecation")
public class explorer_android {
    Fragment frag = new explorer_android_fragment();

    // List of requestCode used in the callback, to identify the caller.
    static List<Integer> import_codes = new ArrayList<>();
    static List<Integer> export_codes = new ArrayList<>();

    // Functions defined on Golang.
    static public native void ImportCallback(FileInputStream f, int id, FileInfo fileInfo, String err);

    static public native void ExportCallback(FileOutputStream f, int id, FileInfo fileInfo, String err);

    public static class FileInfo {
        FileInfo(String uri, String displayName, long size) {
            this.uri = uri;
            this.displayName = displayName;
            this.size = size;
        }

        String uri;
        String displayName;
        long size;
    }

    public static class explorer_android_fragment extends Fragment {
        Context context;

        @Override
        public void onAttach(Context ctx) {
            context = ctx;
            super.onAttach(ctx);
        }

        @Override
        public void onActivityResult(int requestCode, int resultCode, Intent data) {
            super.onActivityResult(requestCode, resultCode, data);

            Activity activity = this.getActivity();
            ContentResolver resolver = activity.getApplicationContext().getContentResolver();

            activity.runOnUiThread(new Runnable() {
                public void run() {
                    if (import_codes.contains(requestCode)) {
                        import_codes.remove(Integer.valueOf(requestCode));
                        if (resultCode != Activity.RESULT_OK) {
                            explorer_android.ImportCallback(null, requestCode, null, "");
                            activity.getFragmentManager().popBackStack();
                            return;
                        }
                        try {
                            Uri uri = Objects.requireNonNull(data.getData());
                            resolver.takePersistableUriPermission(data.getData(), Intent.FLAG_GRANT_READ_URI_PERMISSION);
                            ParcelFileDescriptor pfd = resolver.openFileDescriptor(uri, "r");
                            FileInputStream f = new ParcelFileDescriptor.AutoCloseInputStream(pfd);
                            explorer_android.ImportCallback(f, requestCode, getFileInfoFromContentUri(data.getData()), "");
                        } catch (Exception e) {
                            explorer_android.ImportCallback(null, requestCode, null, e.toString());
                            return;
                        }
                    }

                    if (export_codes.contains(requestCode)) {
                        export_codes.remove(Integer.valueOf(requestCode));
                        if (resultCode != Activity.RESULT_OK) {
                            explorer_android.ExportCallback(null, requestCode, null, "");
                            activity.getFragmentManager().popBackStack();
                            return;
                        }
                        try {
                            Uri uri = Objects.requireNonNull(data.getData());
                            ParcelFileDescriptor pfd = resolver.openFileDescriptor(uri, "wt");
                            FileOutputStream f = new ParcelFileDescriptor.AutoCloseOutputStream(pfd);
                            explorer_android.ExportCallback(f, requestCode, null, "");
                        } catch (Exception e) {
                            explorer_android.ExportCallback(null, requestCode, null, e.toString());
                        }
                    }
                }
            });

        }

        private FileInfo getFileInfoFromContentUri(Uri uri) {
            String[] projection = {OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE};
            try (Cursor cursor = context.getContentResolver().query(uri, projection, null, null, null)) {
                if (cursor != null && cursor.moveToFirst()) {
                    String displayName = cursor.getString(cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME));
                    long size = cursor.getLong(cursor.getColumnIndex(OpenableColumns.SIZE));
                    return new FileInfo(uri.toString(), displayName, size);
                }
            } catch (Exception e) {
                Log.w("explorer", "get file info failed, " + e.getMessage());
            }
            return null;
        }
    }

    public void exportFile(View view, String filename, int id) {
        askPermission(view);

        ((Activity) view.getContext()).runOnUiThread(new Runnable() {
            public void run() {
                registerFrag(view);
                export_codes.add(id);
                int extIndex = filename.lastIndexOf(".") + 1;
                String ext = extIndex < filename.length() ? filename.substring(extIndex) : filename;
                final Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                intent.setType(MimeTypeMap.getSingleton().getMimeTypeFromExtension(ext.toLowerCase()));
                intent.addCategory(Intent.CATEGORY_OPENABLE);
                intent.putExtra(Intent.EXTRA_TITLE, filename);
                frag.startActivityForResult(Intent.createChooser(intent, ""), id);
            }
        });
    }

    public void importFile(View view, String mime, int id) {
        askPermission(view);

        ((Activity) view.getContext()).runOnUiThread(new Runnable() {
            public void run() {
                registerFrag(view);
                import_codes.add(id);
                frag.startActivityForResult(Intent.createChooser(newOpenDocumentIntent(mime), ""), id);
            }
        });
    }

    private static Intent newOpenDocumentIntent(String mime) {
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.setType("*/*");
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION);
        if (mime != null && !mime.isEmpty()) {
            String[] extensions = mime.split(",");
            intent.putExtra(Intent.EXTRA_MIME_TYPES, extensions);
        }
        return intent;
    }

    public FileInputStream openFileInputStream(View view, String uri) {
        askPermission(view);
        Activity activity = (Activity) view.getContext();
        ContentResolver resolver = activity.getApplicationContext().getContentResolver();
        try {
            if (uri.startsWith("content://")) {
                ParcelFileDescriptor pfd = resolver.openFileDescriptor(Uri.parse(uri), "r");
                return new ParcelFileDescriptor.AutoCloseInputStream(pfd);
            } else {
                File file = new File(uri);
                return new FileInputStream(file);
            }
        } catch (IOException e) {
            return null;
        }
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

        Fragment f = fm.findFragmentByTag("explorer_android_fragment");
        // Check if fragment exists and is valid (not removed, not detached, and added)
        if (f != null && f.isAdded() && !f.isRemoving() && !f.isDetached()) {
            return; // Already exists and is valid;
        }

        // Create new fragment instance
        frag = new explorer_android_fragment();

        FragmentTransaction ft = fm.beginTransaction();
        // Replace will automatically remove any existing fragment with the same tag/container
        ft.replace(android.R.id.content, frag, "explorer_android_fragment");
        ft.commitNow();
    }

    public void askPermission(View view) {
        Activity activity = (Activity) view.getContext();

        if (activity.checkSelfPermission(Manifest.permission.READ_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
            activity.requestPermissions(new String[]{Manifest.permission.READ_EXTERNAL_STORAGE}, 255);
        }

        if (activity.checkSelfPermission(Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
            activity.requestPermissions(new String[]{Manifest.permission.WRITE_EXTERNAL_STORAGE}, 254);
        }
    }
}
