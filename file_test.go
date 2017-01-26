package memfs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"bazil.org/fuse"

	"golang.org/x/net/context"

	. "github.com/bbengfort/memfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {

	var ok bool
	var err error
	var tmpDir string
	var config *Config
	var fs *FileSystem
	var root *Dir

	Context("read/write file system", func() {

		BeforeEach(func() {
			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).ShouldNot(HaveOccurred())

			config = makeTestConfig()
			mount := filepath.Join(tmpDir, "testmp")

			fs = New(mount, config)

			node, err := fs.Root()
			Ω(err).ShouldNot(HaveOccurred())
			root, ok = node.(*Dir)
			Ω(ok).Should(BeTrue())
		})

		It("should initialize as a node does", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			Ω(file.Data).ShouldNot(BeZero())
			Ω(file.ID).ShouldNot(BeZero())
			Ω(file.Parent).Should(Equal(root))
			Ω(file.XAttrs).ShouldNot(BeZero())
			Ω(file.Attrs.Inode).ShouldNot(BeZero())
			Ω(file.Attrs.Size).Should(BeZero())
			Ω(file.Attrs.Blocks).Should(BeZero())
			Ω(file.Attrs.Atime).ShouldNot(BeZero())
			Ω(file.Attrs.Mtime).ShouldNot(BeZero())
			Ω(file.Attrs.Ctime).ShouldNot(BeZero())
			Ω(file.Attrs.Crtime).ShouldNot(BeZero())
			Ω(file.Attrs.Mode).Should(Equal(os.FileMode(0644)))
			Ω(file.Attrs.Nlink).Should(Equal(uint32(1)))
			Ω(file.Attrs.Uid).ShouldNot(BeZero())
			Ω(file.Attrs.Gid).ShouldNot(BeZero())
		})

		It("should set the size correctly on setattr", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(4107))
			file.Data = data

			ctx := context.TODO()
			req := &fuse.SetattrRequest{Size: 4107, Valid: fuse.SetattrSize}
			resp := &fuse.SetattrResponse{Attr: file.Attrs}

			err := file.Setattr(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(file.Attrs.Size).Should(Equal(uint64(4107)))
			Ω(file.Attrs.Blocks).Should(Equal(uint64(9)))
			Ω(file.Data).Should(Equal(data))
		})

		It("should truncate data on setattr size", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(4107))
			file.Data = data

			ctx := context.TODO()
			req := &fuse.SetattrRequest{Size: 1056, Valid: fuse.SetattrSize}
			resp := &fuse.SetattrResponse{Attr: file.Attrs}

			err := file.Setattr(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(file.Attrs.Size).Should(Equal(uint64(1056)))
			Ω(file.Attrs.Blocks).Should(Equal(uint64(3)))
			Ω(file.Data).Should(Equal(data[:1056]))
		})

		It("should return all data on read all", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			data := []byte(randString(4107))
			file.Data = data
			file.Attrs.Size = 4107

			ctx := context.TODO()
			resp, err := file.ReadAll(ctx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp).Should(Equal(data))

		})

		It("should return part of the data on read", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			data := []byte(randString(4107))
			file.Data = data
			file.Attrs.Size = 4107

			ctx := context.TODO()
			req := &fuse.ReadRequest{
				Offset: 0, Size: 4107,
			}
			resp := &fuse.ReadResponse{}

			err := file.Read(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(resp.Data).Should(Equal(data))
		})

		It("should return data with an offset", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			data := []byte(randString(4107))
			file.Data = data
			file.Attrs.Size = 4107

			ctx := context.TODO()
			req := &fuse.ReadRequest{
				Offset: 100, Size: 400,
			}
			resp := &fuse.ReadResponse{}

			err := file.Read(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(resp.Data).Should(Equal(data[100:500]))
		})

		It("should return only the data with an incorrect offset", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			data := []byte(randString(4107))
			file.Data = data
			file.Attrs.Size = 4107

			ctx := context.TODO()
			req := &fuse.ReadRequest{
				Offset: 100, Size: 4107,
			}
			resp := &fuse.ReadResponse{}

			err := file.Read(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(resp.Data).Should(Equal(data[100:4107]))
		})

		It("should be able to write data into an empty file", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(4107))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 0,
				Data:   data,
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.Size).Should(Equal(4107))

			Ω(file.Data).Should(Equal(data))
			Ω(file.Attrs.Size).Should(Equal(uint64(4107)))
			Ω(file.Attrs.Blocks).Should(Equal(uint64(9)))
		})

		It("should be able to write then read", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(4107))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 0,
				Data:   data,
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(file.ReadAll(ctx)).Should(Equal(data))

			rreq := &fuse.ReadRequest{
				Offset: 0, Size: 4107,
			}
			rresp := &fuse.ReadResponse{}

			err = file.Read(ctx, rreq, rresp)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(rresp.Data).Should(Equal(data))

		})

		It("should be able to write multiple chunks of data", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(8192))

			ctx := context.TODO()
			for i := int64(0); i < 8192; i += 512 {
				req := &fuse.WriteRequest{
					Offset: i,
					Data:   data[i : i+512],
				}
				resp := &fuse.WriteResponse{}

				err := file.Write(ctx, req, resp)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(resp.Size).Should(Equal(512))
			}

			Ω(file.Data).Should(Equal(data))
			Ω(file.Attrs.Size).Should(Equal(uint64(8192)))
			Ω(file.Attrs.Blocks).Should(Equal(uint64(16)))

			radat, err := file.ReadAll(ctx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(radat).Should(Equal(data))
		})

		It("should be able to overwrite data", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			file.Data = []byte(randString(4107))
			file.Attrs.Size = 4107

			newData := []byte(randString(4107))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 0,
				Data:   newData,
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(file.Data).Should(Equal(newData))
			Ω(file.Attrs.Size).Should(Equal(uint64(4107)))

		})

		It("should be able to truncate data on write", func() {
			Skip("should it not allow truncation?")

			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			file.Data = []byte(randString(4107))
			file.Attrs.Size = 4107

			newData := []byte(randString(1852))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 0,
				Data:   newData,
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(file.Data).Should(Equal(newData))
			Ω(file.Attrs.Size).Should(Equal(uint64(1852)))
		})

		It("should be able to update portions of data", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			file.Data = []byte("the cat in the hat sat on the bat")
			file.Attrs.Size = uint64(len(file.Data))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 19,
				Data:   []byte("ran across the mat until he was very tired"),
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(file.Data).Should(Equal([]byte("the cat in the hat ran across the mat until he was very tired")))
			Ω(file.Attrs.Size).Should(Equal(uint64(61)))
		})

	})

	Context("read only file system", func() {

		BeforeEach(func() {
			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).ShouldNot(HaveOccurred())

			config = makeTestConfig()
			config.ReadOnly = true
			mount := filepath.Join(tmpDir, "testmp")

			fs = New(mount, config)

			node, err := fs.Root()
			Ω(err).ShouldNot(HaveOccurred())
			root, ok = node.(*Dir)
			Ω(ok).Should(BeTrue())
		})

		It("should not allow Setattr", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(4107))
			file.Data = data

			ctx := context.TODO()
			req := &fuse.SetattrRequest{Size: 4107, Valid: fuse.SetattrSize}
			resp := &fuse.SetattrResponse{Attr: file.Attrs}

			err := file.Setattr(ctx, req, resp)
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(fuse.EPERM))
		})

		It("should not alow Flush", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)

			ctx := context.TODO()
			req := &fuse.FlushRequest{}
			err := file.Flush(ctx, req)
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(fuse.EPERM))
		})

		It("should not allow Write", func() {
			file := new(File)
			file.Init("test.txt", 0644, root, fs)
			data := []byte(randString(3264))

			ctx := context.TODO()
			req := &fuse.WriteRequest{
				Offset: 0,
				Data:   data,
			}
			resp := &fuse.WriteResponse{}

			err := file.Write(ctx, req, resp)
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(Equal(fuse.EPERM))
		})

	})

})
