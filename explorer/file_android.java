package org.gioui.x.explorer;

import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.Closeable;
import java.io.Flushable;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;

public class file_android {
    public String err;
    public Object handler;
    private FileChannel fileChannel;

    public void setHandle(Object f) {
        this.handler = f;
        if (this.handler instanceof FileInputStream) {
            this.fileChannel = ((FileInputStream) this.handler).getChannel();
        }
        if (this.handler instanceof FileOutputStream) {
            this.fileChannel = ((FileOutputStream) this.handler).getChannel();
        }
    }

    public int fileRead(byte[] b) {
        try {
            return this.fileChannel.read(ByteBuffer.wrap(b));
        } catch (Exception e) {
            this.err = e.toString();
            return 0;
        }
    }

    public boolean fileWrite(byte[] b) {
        try {
            return this.fileChannel.write(ByteBuffer.wrap(b)) > 0;
        } catch (Exception e) {
            this.err = e.toString();
            return false;
        }
    }

    public boolean filePosition(int position) {
        try {
            this.fileChannel.position(position);
            return true;
        } catch (Exception e) {
            this.err = e.toString();
            return false;
        }
    }

    public boolean fileClose() {
        try {
            if (this.handler instanceof Flushable) {
                ((Flushable) this.handler).flush();
            }
            if (this.handler instanceof Closeable) {
                ((Closeable) this.handler).close();
            }
            return true;
        } catch (Exception e) {
            this.err = e.toString();
            return false;
        }
    }

    public String getError() {
        return this.err;
    }

}