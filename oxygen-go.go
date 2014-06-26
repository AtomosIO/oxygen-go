package oxygen

import (
	//"bufio"
	"fmt"
	"io"
	"net/url"
)

// Clients must be safe to access from concurrent go routines
type Client interface {
	// Resolve the attributes of a node starting from startNode
	ResolvePathFromNode(startNode int64, path string) (attr *NodeAttributes, err error)
	// Similiar to ResolvePathFromNode but assumes root as starting node
	ResolvePath(path string) (attr *NodeAttributes, err error)
	// Resolve the attributes of a node
	ResolveNode(node int64) (attr *NodeAttributes, err error)

	// Read the data in the nodeId starting at offset.
	// A size of -1 signifies no limit on size.
	ReadNode(nodeId int64, offset int64, size int) (attr *NodeAttributes, data io.ReadCloser, err error)
	// Similiar to ReadNode but using a path instead of a nodeId. A size of -1
	// signifies no limit on size.
	ReadPath(path string, offset int64, size int) (attr *NodeAttributes, data io.ReadCloser, err error)

	// Replace the data of the file. Current Limitation: Offset MUST be 0.
	OverwriteNode(nodeId int64, offset int64, data io.Reader) (attr *NodeAttributes, err error)
	// Create or replace the data of the file. Current Limitation: Offset must be 0.
	OverwritePath(path string, offset int64, data io.Reader) (attr *NodeAttributes, err error)
	// Create or replace the data of the file. Current Limitation: Offset must be 0.
	OverwritePathFromNode(nodeId int64, path string, offset int64, data io.Reader) (attr *NodeAttributes, err error)

	// Create a file/directory if it does not exist. If the file exists, will return an error
	CreatePathFromNode(nodeId int64, path string, data io.Reader) (attr *NodeAttributes, err error)

	CreatePath(path string, data io.Reader) (attr *NodeAttributes, err error)

	// Delete a file from the specified node
	DeleteFromNode(nodeId int64, entry string) (err error)

	// Delete a file from the specified node
	RenameFromNodeToNode(oldNodeId int64, oldName string, newNodeId int64, newName string) (err error)

	Logf(format string, args ...interface{})
}

type NodeAttributes struct {
	Id   int64
	Type byte
	Size int64
}

/*type bufferedReadCloser struct {
	bufReader  *bufio.Reader
	readCloser io.ReadCloser
}

func newBufferedReadCloser(readCloser io.ReadCloser) *bufferedReadCloser {
	return &bufferedReadCloser{
		readCloser: readCloser,
		bufReader:  bufio.NewReader(readCloser),
	}
}

func (bufferedReadCloser *bufferedReadCloser) Close() error {
	return bufferedReadCloser.readCloser.Close()
}

func (bufferedReadCloser *bufferedReadCloser) Read(p []byte) (n int, err error) {
	return bufferedReadCloser.bufReader.Read(p)
}
*/

type URL struct {
	url.URL
}

func (client *HttpClient) NewURL(format string, args ...interface{}) *URL {
	return &URL{
		URL: url.URL{
			Scheme: client.scheme,
			Host:   client.host,
			Path:   fmt.Sprintf(format, args...),
		},
	}
}

func (urlVar *URL) SetIdQuery() *URL {
	return urlVar.AddStringToQuery("id=true")
}

func (urlVar *URL) SetOverwriteQuery() *URL {
	return urlVar.AddStringToQuery("overwrite=true")
}

func (urlVar *URL) AddStringToQuery(query string) *URL {
	if urlVar.RawQuery == "" {
		urlVar.RawQuery = query
	} else {
		urlVar.RawQuery += "&" + query
	}
	return urlVar
}
