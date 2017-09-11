package gopathfs

import (
	"path/filepath"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"golang.org/x/sys/unix"
)

// GetAttr overwrites the parent's GetAttr method.
func (gpf *GoPathFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return gpf.getTopDirAttr()
	}

	// Handle the virtual Golang prefix package.
	if name == gpf.cfg.GoPkgPrefix {
		return gpf.getFirstPartyDirAttr()
	}

	// Handle the children of the virtual Golang prefix package.
	prefix := gpf.cfg.GoPkgPrefix + "/"
	if strings.HasPrefix(name, prefix) {
		name = name[len(prefix):]
		attr, status := gpf.getFirstPartyChildDirAttr(name)
		if status == fuse.OK {
			return attr, fuse.OK
		}
	}

	// Search in vendor directories.
	for _, v := range gpf.cfg.Vendors {
		fname := filepath.Join(gpf.dirs.Workspace, v, name)
		attr, status := gpf.getRealDirAttr(fname)
		if status == fuse.OK {
			return attr, fuse.OK
		}

		// Also search in bezel-genfiles.
		fname = filepath.Join(gpf.dirs.Workspace, "bazel-genfiles", v, name)
		attr, status = gpf.getRealDirAttr(fname)
		if status == fuse.OK {
			return attr, fuse.OK
		}
	}

	return nil, fuse.ENOENT
}

func (gpf *GoPathFs) getTopDirAttr() (*fuse.Attr, fuse.Status) {
	return &fuse.Attr{
		Mode: fuse.S_IFDIR | 0755,
	}, fuse.OK
}

func (gpf *GoPathFs) getFirstPartyDirAttr() (*fuse.Attr, fuse.Status) {
	return &fuse.Attr{
		Mode: fuse.S_IFDIR | 0755,
	}, fuse.OK
}

func (gpf *GoPathFs) getFirstPartyChildDirAttr(name string) (*fuse.Attr, fuse.Status) {
	nm := filepath.Join(gpf.dirs.Workspace, name)
	attr, status := gpf.getRealDirAttr(name)
	if status == fuse.OK {
		return attr, fuse.OK
	}

	// Search in bazel-genfiles directories.
	nm = filepath.Join(gpf.dirs.Workspace, "bazel-genfiles", name)
	return gpf.getRealDirAttr(nm)
}

func (gpf *GoPathFs) getRealDirAttr(name string) (*fuse.Attr, fuse.Status) {
	t := unix.Stat_t{}
	err := unix.Stat(name, &t)
	if err != nil {
		return nil, fuse.ENOENT
	}

	attr := &fuse.Attr{
		Ino:    t.Ino,
		Size:   uint64(t.Size),
		Blocks: uint64(t.Blocks),
		Mode:   uint32(t.Mode),
	}

	sec, nsec := t.Atimespec.Unix()
	attr.Atime = uint64(sec)
	attr.Atimensec = uint32(nsec)

	sec, nsec = t.Ctimespec.Unix()
	attr.Ctime = uint64(sec)
	attr.Ctimensec = uint32(nsec)

	sec, nsec = t.Mtimespec.Unix()
	attr.Mtime = uint64(sec)
	attr.Mtimensec = uint32(nsec)

	return attr, fuse.OK
}
