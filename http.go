package oxygen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	NODE_ID_HEADER_KEY   = "Node-Id"
	NODE_TYPE_HEADER_KEY = "Node-Type"
	NODE_SIZE_HEADER_KEY = "Node-Size"
)

const (
	INVALID = iota
	DIRECTORY
	FILE
)

type HttpClient struct {
	endpoint string
	token    string
	cache    Cache
	client   *http.Client
	log      bool

	//URL Related data
	scheme string
	host   string
	path   string
}

type PatchParameters struct {
	Path        string `json:"path"`
	PathUsingId bool   `json:"path_using_id"`
}

var (
	ErrWriteOffsetNotSupported = errors.New("Non-zero offset values are not currently supported for write operations")
	ErrRangeNotSatisfiable     = errors.New("Range not satisfiable")
	ErrDidNotSucceed           = errors.New("HTTP Request returned none 2XX status code")
	ErrUnableToResolvePathId   = errors.New("Unable to resolve path with id")
	ErrInvalidNodeIdString     = errors.New("Unable to convert node id string to integer")
	ErrInvalidNodeTypeString   = errors.New("Unable to convert node type string to integer")
	ErrNotEnoughPermissions    = errors.New("Not enough permissions to perform task")
	ErrDirectoryNotEmpty       = errors.New("Directory not empty")
)

// Create a new Oxygen client using HTTP protocol. The client will use the
// specified token for all interactions with the service. An empty token string
// will cause the omission of token cookie, resulting in only public repositories
// being accesible.
func NewHttpClient(endpoint, token string) (client *HttpClient) {
	return newHttpClient(endpoint, token, false)
}

func newHttpClient(endpoint, token string, log bool) (client *HttpClient) {
	urlURL, err := url.Parse(endpoint)
	if err != nil {
		panic(err)
	}

	return &HttpClient{
		token:    token,
		endpoint: endpoint,
		client:   &http.Client{},
		log:      log,

		scheme: urlURL.Scheme,
		host:   urlURL.Host,
		path:   urlURL.Path,
	}
}

func (client *HttpClient) StartLogging() *HttpClient {
	client.log = true
	return client
}

func (client *HttpClient) Logf(format string, args ...interface{}) {
	if client.log {
		fmt.Printf(format, args...)
	}
}

// Set the cache policy for the client.
func (client *HttpClient) SetCache(cache Cache) {
	client.cache = cache
}

func (client *HttpClient) prepHeadRequest(url *URL) (req *http.Request, err error) {
	return client.prepEmptyRequest("HEAD", url)
}

func (client *HttpClient) prepDeleteRequest(url *URL) (req *http.Request, err error) {
	return client.prepEmptyRequest("DELETE", url)
}

func (client *HttpClient) prepGetRequest(url *URL) (req *http.Request, err error) {
	return client.prepEmptyRequest("GET", url)
}

func (client *HttpClient) prepEmptyRequest(method string, url *URL) (req *http.Request, err error) {
	return client.prepRequest(method, url, NewEmptyReader())
}

func (client *HttpClient) prepPostRequest(url *URL, body io.Reader) (req *http.Request, err error) {
	return client.prepRequest("POST", url, body)
}

func (client *HttpClient) prepPatchRequest(url *URL, body io.Reader) (req *http.Request, err error) {
	return client.prepRequest("PATCH", url, body)
}

func (client *HttpClient) prepRequest(method string, urlString *URL, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, urlString.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", client.token)
	/*if client.token != "" {
		tokenCookie := &http.Cookie{
			Name:  "token",
			Value: client.token,
		}
		req.AddCookie(tokenCookie)
	}*/

	//client.Logf("HttpClient %s -> %s\n", method, url)

	return req, nil
}

func (client *HttpClient) ResolvePathFromNode(startNode int64, path string) (attr *NodeAttributes, err error) {
	var url *URL
	if path == "" {
		url = client.NewURL("%d", startNode).SetIdQuery()
	} else {
		url = client.NewURL("%d/%s", startNode, path).SetIdQuery()
	}

	// Prepare request
	req, err := client.prepHeadRequest(url)
	if err != nil {
		client.Logf("Failed PrepHeadRequest: %s\n", err)
		return nil, err
	}

	// Do request
	attr, body, err := client.doRequest(req)

	// Close body reader
	if body != nil {
		body.Close()
	}

	return attr, err
}

func (client *HttpClient) ResolvePath(path string) (attr *NodeAttributes, err error) {
	return client.ResolvePathFromNode(1, path)
}

func (client *HttpClient) ResolveNode(nodeId int64) (attr *NodeAttributes, err error) {
	return client.ResolvePathFromNode(nodeId, "")
}

func (client *HttpClient) ReadNode(nodeId int64, offset int64, size int) (attr *NodeAttributes, data io.ReadCloser, err error) {
	url := client.NewURL("%d", nodeId).SetIdQuery()
	return client.read(url, offset, size)
}

func (client *HttpClient) ReadPath(path string, offset int64, size int) (attr *NodeAttributes, data io.ReadCloser, err error) {
	url := client.NewURL("%s", path)
	return client.read(url, offset, size)
}

func (client *HttpClient) OverwriteNode(nodeId int64, offset int64, data io.Reader) (attr *NodeAttributes, err error) {
	url := client.NewURL("%d", nodeId).SetIdQuery().SetOverwriteQuery()
	return client.write(url, offset, data)
}

func (client *HttpClient) OverwritePath(path string, offset int64, data io.Reader) (attr *NodeAttributes, err error) {
	url := client.NewURL("%s", path).SetOverwriteQuery()
	return client.write(url, offset, data)
}

func (client *HttpClient) OverwritePathFromNode(nodeId int64, path string, offset int64, data io.Reader) (attr *NodeAttributes, err error) {
	url := client.NewURL("%d/%s", nodeId, path).SetIdQuery().SetOverwriteQuery()
	return client.write(url, offset, data)
}

func (client *HttpClient) CreatePathFromNode(nodeId int64, path string, data io.Reader) (attr *NodeAttributes, err error) {
	url := client.NewURL("%d/%s", nodeId, path).SetIdQuery()
	return client.write(url, 0, data)
}

func (client *HttpClient) CreatePath(path string, data io.Reader) (attr *NodeAttributes, err error) {
	url := client.NewURL("%s", path)
	return client.write(url, 0, data)
}

func (client *HttpClient) DeleteFromNode(nodeId int64, entry string) (err error) {
	url := client.NewURL("%d/%s", nodeId, entry).SetIdQuery()
	return client.delete(url)
}

func (client *HttpClient) RenameFromNodeToNode(oldNodeId int64, oldName string, newNodeId int64, newName string) (err error) {
	url := client.NewURL("%d/%s", oldNodeId, oldName).SetIdQuery().SetOverwriteQuery()
	parameters := &PatchParameters{
		Path:        fmt.Sprintf("/%d/%s", newNodeId, newName),
		PathUsingId: true,
	}
	body, _ := json.Marshal(parameters)
	return client.patch(url, body)
}

func (client *HttpClient) read(url *URL, offset int64, size int) (attr *NodeAttributes, data io.ReadCloser, err error) {
	// Prepare request
	req, err := client.prepGetRequest(url)
	if err != nil {
		client.Logf("Failed PrepRequest: %s\n", err)
		return nil, nil, err
	}
	// If we defined an offset or size, set the HTTP Range header
	setRequestRangeHeader(req, offset, size)

	// Do request
	attr, body, err := client.doRequest(req)

	return attr, body, err
}

func (client *HttpClient) write(url *URL, offset int64, reader io.Reader) (attr *NodeAttributes, err error) {
	if offset != 0 {
		return nil, ErrWriteOffsetNotSupported
	}

	// Prepare request
	req, err := client.prepPostRequest(url, reader)
	if err != nil {
		client.Logf("Failed PrepPostRequest: %s\n", err)
		return nil, err
	}

	// Do request
	attr, body, err := client.doRequest(req)

	// Close body reader
	if body != nil {
		body.Close()
	}

	return attr, err
}

func (client *HttpClient) patch(url *URL, buf []byte) (err error) {
	// Prepare request
	req, err := client.prepPatchRequest(url, bytes.NewReader(buf))
	if err != nil {
		client.Logf("Failed prepPatchRequest: %s\n", err)
		return err
	}

	// Do request
	_, body, err := client.doRequest(req)

	// Close body reader
	if body != nil {
		body.Close()
	}

	return err
}

func (client *HttpClient) delete(url *URL) (err error) {

	// Prepare request
	req, err := client.prepDeleteRequest(url)
	if err != nil {
		client.Logf("Failed prepDeleteRequest: %s\n", err)
		return err
	}

	// Do request
	_, body, err := client.doRequest(req)

	// Close body reader
	if body != nil {
		body.Close()
	}

	return err
}

func (client *HttpClient) doRequest(req *http.Request) (attr *NodeAttributes, readCloser io.ReadCloser, err error) {
	// Do request
	if client.log {
		body, _ := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	resp, err := client.client.Do(req)
	if err != nil {
		client.Logf("Failed Do: %s\n%+v\n", err, req)
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		return nil, nil, err
	}

	if client.log {
		client.Logf("%+v\n", resp)
		body, _ := ioutil.ReadAll(resp.Body)
		client.Logf("%s\n", body)
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	if !statusGood(resp.StatusCode) {
		return nil, nil, ParseErrorResponse(resp)
	}

	// Parse headers for node attributes
	attr, err = client.parseNodeAttributes(resp)
	//fmt.Println("node attr", attr, err)
	if err != nil {
		client.Logf("Failed ParseNodeAttributesf: %s", err)
		if resp.Body != nil {
			resp.Body.Close()
		}
		return nil, nil, err
	}

	return attr, resp.Body, nil
}

func ParseErrorResponse(resp *http.Response) error {
	oxygenResponse := &OxygenResponse{}
	buf, _ := ioutil.ReadAll(resp.Body)
	if buf != nil {
		json.Unmarshal(buf, oxygenResponse)
	}

	switch resp.StatusCode {
	case http.StatusForbidden:
		return ParseForbiddenErrorResponse(oxygenResponse)
	case http.StatusRequestedRangeNotSatisfiable:
		return ErrRangeNotSatisfiable
	default:
		/*if resp.StatusCode != http.StatusNotFound {
			fmt.Printf("%+v\n%s\n%s\n", resp, buf, err)
		}*/
		return ErrDidNotSucceed
	}
}

func ParseForbiddenErrorResponse(oxygenResponse *OxygenResponse) error {
	switch oxygenResponse.Code {
	case ERROR_DIRECTORY_NOT_EMPTY:
		return ErrDirectoryNotEmpty
	default:
		return ErrNotEnoughPermissions
	}
}

func (client *HttpClient) parseNodeAttributes(resp *http.Response) (attr *NodeAttributes, err error) {
	// Parse Node Id
	nodeIdString := resp.Header.Get(NODE_ID_HEADER_KEY)
	if nodeIdString == "" {
		client.Logf("ResolvePathId no nodeIdString: %s\n", resp)
		return nil, ErrUnableToResolvePathId
	}

	nodeId, err := strconv.ParseInt(nodeIdString, 10, 64)
	if err != nil {
		client.Logf("ResolvePathId failed to parse nodeId %s: %s\n", nodeIdString, err)
		return nil, ErrInvalidNodeIdString
	}

	// Parse Node Size
	nodeSizeString := resp.Header.Get(NODE_SIZE_HEADER_KEY)
	nodeSize, _ := strconv.ParseInt(nodeSizeString, 10, 64)

	// Parse Node String
	nodeTypeString := resp.Header.Get(NODE_TYPE_HEADER_KEY)
	var nodeType byte
	switch nodeTypeString {
	case "directory":
		nodeType = DIRECTORY
	case "file":
		nodeType = FILE
	}

	return &NodeAttributes{
		Id:   nodeId,
		Type: nodeType,
		Size: nodeSize,
	}, nil

}

// Did we get a 2XX respond code?
func statusGood(status int) bool {
	return status >= 200 && status <= 299
}

func setRequestRangeHeader(req *http.Request, offset int64, size int) {
	if offset != 0 || size != -1 {
		rangeValue := "bytes=" + strconv.FormatInt(offset, 10) + "-"
		if size != -1 {
			rangeValue += strconv.FormatInt(offset+int64(size)-1, 10)
		}
		req.Header.Add("Range", rangeValue)
	}
}
