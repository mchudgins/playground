package htmlGen

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type staticFilesFile struct {
	data  string
	mime  string
	mtime time.Time
	// size is the size before compression. If 0, it means the data is uncompressed
	size int
	// hash is a sha256 hash of the file contents. Used for the Etag, and useful for caching
	hash string
}

var staticFiles = map[string]*staticFilesFile{
	"apiList.html": {
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccW\xefo\xdb6\x10\xfd\x1c\xff\x157\r\x1bZ`\xfa\xe1l\xc5\x02E\xd6\xe0f)P`\x05\x02$C\xf7-\xa0ȓą\"5\xf2d\xcf3\xf6\xbf\x0f\xb2d[\x8e\x1d'.\x1a`\x9f(\x9f\xdf;\x1e\xdfݣ\xec\xe4\x1ba8-j\x84\x92*\x95\x8e\x92v\x01\xc5t1\xf1P{m\x00\x99HG\x00I\x85Ā\x97\xcc:\xa4\x89\xd7P\xee_x\xdb/J\xa2\xdaǿ\x1a9\x9bx\u007f\xf8\xbfO\xfd+SՌd\xa6\xd0\x03n4\xa1\xa6\x89\xf7\xf1z\x82\xa2\xc0\x1fxiM\x85\x93q\x97\x80$)L\u007f\xbd\xbd\x83\xe9\xcdG\xb8\xfe\xbbVƢM\xc2.\xde\"\x1c-\xba\xa7\xb3r\fK\xc8\x18\u007f(\xaci\xb4\xf0\xb9Q\xc6\xc6\xf0m\x1e\xe5Q\x9e_B̈́\x90\xba\xf0\xc9\xd41\x8c\xb1\xdaF2Cd\xaa>\xf8o\x9bK\xc8Y`\xcd\x1c\x96P1[H\xbd\x81D\xc1\xcf\xef\x86\xd4U\xb2\xe0\xfcݡt}x\x980\xd6T\xfa\xbc\x94J\xbc1B\xbc=X\U0003c504\xae2\x0f\xf8$\x17g\xa8\x8f\x90[^K\f\x1a\xcdfL*\x96)\xbcwĨq\xb0\x84\x1eYX\xb6\xe87\xd8\xc1\tt\xbc\a\xc71\xcb\t튳jS\f\x1e\xbc\x19\x80\xdfz\xeb\f\x8c\x93\x9c\xedo\x92\xa9f}\x8a@`m\x913Bq\xb4\x96\x01\xec\xd9R\xb6\xd8A%\xb5\xbcE\xba\u05ec\xc2{n\x14,AHW+\xb6\x88Aj%5\xfa\x992\xfc\xe1\x11|\xb5\xd5\v\xe0ϧ\x9dKAe\f?mZ\x1f\xcc\xd0:i\xf4\xa9\xb4\xe7K\xea9?F[\x12\xab\xe5\xb5\x16\xb5\x91\x9a\x86\xbc\x9e\xd0\x0f\xb3\u009cv\xb6\xeaD\x80%\xe4F\x93\xef\xe4?\x18\xc3x\x1c}w\xd9\x05\xe6(\x8b\x92bȌ\x12\x97\x8f\x1d1\x98\xf2$\xecݘ\x84\xddݐdF,ڛb\x9c\xee\xfa\xb7\x1c\xa7\xa3D\xc8Y:Z-\xc0\x15sn\xe2Y3\xf7\xd2\xd1\xd90ԕ\xd6F\x0f\x847]\xf6\x1e\xe5oS\u007f\xaf3W_\xc6\xddr\x80\xbb\xd6\xd7K?1\xcd\n\x04*\x11\xf6ӌ\xce\xd6\xeb^a\x1b\xa5w\xa6\u007f\xaf\xd6m\x91\t\x83\xd2b>\xf1\xda;1\x0eCe8S\xa5q\x14_D\x17Q\xe8\xe6\xac(\xd0\xfa\x8d\f\u007fi\xac\x9a\x1cE\x85\x0e\xedLr\f\xfa\xcf\xc1\x9f\xceh/݀\x93\x90\xa5\xeb\xcaw\xca\x19\x8c\xa3\x97\xce\xc6\a1\x9b\xd9\xebO6p\xa2\x97ޡ#\x90\xda\x11\xd3\x1c\xc16ZK]\x80Ѱ0\x8d\x85\x8a\xf1Rj|}\xf12Y\xb4\xd7K\xa0\x98>*\xdfS\xb8'\x04\xac\xad\x11\xaf\xa7\xddM\x93)ɏ\x8fY\xff\xf0\xb5\x9cqu\xfb\xf9TC\\5\x8eL\x85\x16n;\x89ೱ\x0f\xed\x11\xa4\xd1/o\xecޅ\u007fBs\xab\x9c\x05\xc2\x117\xb6\x0e4R\xf85\xfbr\xf8\r㥿a\xc1\xf8\x02ڍ\x1a\xde\x1eu3\xe4\xafݢ\xbb\xe9y\x14E'_[\r5L\xc1\x87F\v\xb8\xb3L\xbb\x1c-L\v\xd4|\x01\xb7\vGX\xbd\xbe\t\xd7\xde\x19\xf6jh\xbe\xc7}<\xee\xbdO\x1f\xa6/l\xf19Suɾ\xc0\x80Ӗ\a\xab+\xecF1ʍ\xfd_\xa8\x94\x9d\xa4\xd2\xfb\x97\x1a!\xf8\x02\x85\u07b7?ڏ\v\xb4\x19\xffnM\xc2\xfe=\x1f\xae\xfe*\xfc\x17\x00\x00\xff\xff\xf3\x96\xb4\x01:\f\x00\x00",
		hash:  "1c5a7acebcdd253bfbb3613e20384c943a09847268a80b14a0b59ea0770effbf",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1489529044, 0),
		size:  3130,
	},
}

// NotFound is called when no asset is found.
// It defaults to http.NotFound but can be overwritten
var NotFound = http.NotFound

// ServeHTTP serves a request, attempting to reply with an embedded file.
func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f, ok := staticFiles[strings.TrimPrefix(req.URL.Path, "/")]
	if !ok {
		NotFound(rw, req)
		return
	}
	header := rw.Header()
	if f.hash != "" {
		if hash := req.Header.Get("If-None-Match"); hash == f.hash {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("ETag", f.hash)
	}
	if !f.mtime.IsZero() {
		if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && f.mtime.Before(t.Add(1*time.Second)) {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("Last-Modified", f.mtime.UTC().Format(http.TimeFormat))
	}
	header.Set("Content-Type", f.mime)

	// Check if the asset is compressed in the binary
	if f.size == 0 {
		header.Set("Content-Length", strconv.Itoa(len(f.data)))
		io.WriteString(rw, f.data)
	} else {
		if header.Get("Content-Encoding") == "" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			header.Set("Content-Encoding", "gzip")
			header.Set("Content-Length", strconv.Itoa(len(f.data)))
			io.WriteString(rw, f.data)
		} else {
			header.Set("Content-Length", strconv.Itoa(f.size))
			reader, _ := gzip.NewReader(strings.NewReader(f.data))
			io.Copy(rw, reader)
			reader.Close()
		}
	}
}

// Server is simply ServeHTTP but wrapped in http.HandlerFunc so it can be passed into net/http functions directly.
var Server http.Handler = http.HandlerFunc(ServeHTTP)

// Open allows you to read an embedded file directly. It will return a decompressing Reader if the file is embedded in compressed format.
// You should close the Reader after you're done with it.
func Open(name string) (io.ReadCloser, error) {
	f, ok := staticFiles[name]
	if !ok {
		return nil, fmt.Errorf("Asset %s not found", name)
	}

	if f.size == 0 {
		return ioutil.NopCloser(strings.NewReader(f.data)), nil
	}
	return gzip.NewReader(strings.NewReader(f.data))
}

// ModTime returns the modification time of the original file.
// Useful for caching purposes
// Returns zero time if the file is not in the bundle
func ModTime(file string) (t time.Time) {
	if f, ok := staticFiles[file]; ok {
		t = f.mtime
	}
	return
}

// Hash returns the hex-encoded SHA256 hash of the original file
// Used for the Etag, and useful for caching
// Returns an empty string if the file is not in the bundle
func Hash(file string) (s string) {
	if f, ok := staticFiles[file]; ok {
		s = f.hash
	}
	return
}
