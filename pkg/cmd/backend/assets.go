package backend

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
	"service.swagger.json": {
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccUOo\xd4>\x10\xbd\xe7S\x8c\xfc\xfbI\\\xa0i{\xdc\x1b\xa2\x12T\xdcZ8\x95=\x98d6q\x95\xd8\ue333h\xa9\xf2ݑ\x9d\u007f\x8e\xbbZ\x10\xb4\x82\xbdl\xe2\x19{\xde{\xf3<y\xcc\x00\x04\u007f\x93U\x85$6 .\xcf\xce\xc5k\xbf\xa6\xf4Έ\r\xf88\x80p\xca5\xe8㌴W\x05\x9eY2΄L\x00\xb1Gbe\xb4\x8f\xef/D\x06Ї#\xb8\xa8\xb1E\x16\x1b\xb8\x1b\xf2j\xe7\xec\xb4\xc7?\xb3\xcf݆\xdc\xc2h\xeeV\xc9\xd2\xdaF\x15\xd2)\xa3\xf3{6zɵdʮ\xf8\xb5\\Ƣ#\xe5\x0eW\xb8SZ\xf9\x04\x1eY\x89όt;\x86#\xa6\a\x1b\x88J\xab>\xe2aB\xab\x02\xb9\x1ae\x894\xadiن̷\x9d\xab\r\xa9\xef\xa1|`\x9f\xf5\xab\xda3\xceǤ\xe6ݶ_HIW\xf3\x02#\x97V\xe5\xfb\x8b\x1c\x8b\xda䏾T?\a\x01D\x85.z\xf5\xb5\xba\xb6\x95\xe4K\x89[\xd4%\x83\x84\x8a\x10\x9d\xd2Ո7\xa4\x19\x8b\x14p^\x97!U\x1e>`Ә8\x85\x90\xadь\xbc*\x00 .\xcfϓ%\x00Q\"\x17\xa4\xac\x1b\xdb\x1f\x1d4\xc0\xf2\x16\x90O\xb6\x01\x88\xff\tw~\xc7\u007fy\xb9\xb4&\x1f\xed\x15@ݠm\x0eb\xb5\xafώ=\xf7\x11z+I\xb6\xe8\x90\x16w\f\xbf\x04\xf7Լ\xf0\x9f\x80\x1e\x9a\xed\x1b\x92F\b\x1f:E\xe8\xb5s\xd4a\x12\x9d\xacÎֲ\x87\xe8\xceP+]\x14?\xcae\x1bqq\xb2JY\x88\xf7\xbe\xa9H\xcb\xe6m\x16\x1f\xd1ϗ\xaf|bx\x98o\xef;ӶF\xdf\xe0C\x87\x1c\xfbhf`\xbe\xdec\xe1f\x06\xfe\xc6Y$\xa7\x12S\x88\xc2\x10a3\x18\xea*\xf5\xcb\t9N\x88\x117\x93\xad|\x91s;\x0e\xc3\xee9N͒ӗI\xf9\xa9F\xa0Abh\x91YV\b\x85\xd1N*\xadt\xb5\xf9\xa2\x01ހ\xab\x11d\xe7j\xd4\xce\xcf/,\xc1c\x83\xeb\xab%\xbc\xd2xY\x1e\xb5ɢ\xeai{\x87{\xfc\xaf\xf5\xf7g\x8a\r\xa8'Ɏ\x12\x8c\xc6\xc3\x1f\x91\xf32\xa5\xacNΥD٣暀\xff\r\xb5\"\x83\x05\x93L\x9f\x00>%\xe33L\x81\xdf\x16r\xa8}T\xc7qB\xbf\x88\x88\xc9W\xeb\xf4U\rJ\xfak\xf9\x8a\xc1\x83:\x13˨\xcd\xfa\xecG\x00\x00\x00\xff\xff\x840\x1c\x15B\t\x00\x00",
		hash:  "04d17d8e87d26ed1a31a963cd03a2a55a9c446a10d86bce7d20a1a60fd2c3c2f",
		mime:  "application/json",
		mtime: time.Unix(1497071481, 0),
		size:  2370,
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
