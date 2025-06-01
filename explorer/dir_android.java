package org.gioui.x.explorer;

import android.content.Context;
import android.net.Uri;
import android.provider.DocumentsContract;
import android.database.Cursor;

import android.view.View;
import java.io.InputStream;
import java.io.OutputStream;

public class dir_android {
    public Object handler;

    public void setHandle(Object f) {
        this.handler = f;
    }

    public Object readFile(View view, String filename) {
        try {
            Context ctx = view.getContext();
            Uri folderUri = (Uri) handler;
            Uri fileUri = findFile(ctx, folderUri, filename);
            if (fileUri == null) {
                return null;
            }

            return ctx.getContentResolver().openInputStream(fileUri);
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    public Object writeFile(View view, String filename) {
        try {
            Context ctx = view.getContext();
            Uri folderUri = (Uri) handler;
            Uri fileUri = findFile(ctx, folderUri, filename);

            if (fileUri == null) {
                fileUri = DocumentsContract.createDocument(ctx.getContentResolver(), folderUri, "application/octet-stream", filename);
            }

            return ctx.getContentResolver().openOutputStream(fileUri, "wt");
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }

    private Uri findFile(Context ctx, Uri folderUri, String filename) {
        String folderDocId = DocumentsContract.getDocumentId(folderUri);
        Uri childrenUri = DocumentsContract.buildChildDocumentsUriUsingTree(folderUri, folderDocId);

        try (Cursor cursor = ctx.getContentResolver().query(childrenUri, new String[] {DocumentsContract.Document.COLUMN_DOCUMENT_ID, DocumentsContract.Document.COLUMN_DISPLAY_NAME},null, null, null)) {
            if (cursor != null) {
                while (cursor.moveToNext()) {
                    String docId = cursor.getString(0);
                    String name = cursor.getString(1);
                    if (filename.equals(name)) {
                        return DocumentsContract.buildDocumentUriUsingTree(folderUri, docId);
                    }
                }
            }
        } catch (Exception e) {
            e.printStackTrace();
        }

        return null;
    }
}
