//go:build !android && !darwin

package explorer

type File struct {
}

func (f *File) Read(b []byte) (n int, err error) {
	return 0, ErrNotAvailable
}
func (f *File) Write(b []byte) (n int, err error) {
	return 0, ErrNotAvailable
}
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return 0, ErrNotAvailable
}
func (f *File) Close() error {
	return ErrNotAvailable
}
func (f *File) Name() string { return "" }
func (f *File) Size() int64  { return 0 }
func (f *File) URI() string  { return "" }
