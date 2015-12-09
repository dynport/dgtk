package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	b64    = base64.StdEncoding
	null   = struct{}{}
	client = &Client{}
)

const (
	minPartSize                = 5 * 1024 * 1024
	headerContentMd5           = "Content-Md5"
	headerContentType          = "Content-Type"
	headerDate                 = "Date"
	headerAuthorization        = "Authorization"
	defaultContentType         = "application/octet-stream"
	headerAmzAcl               = "x-amz-acl"
	headerServerSideEncruption = "x-amz-server-side-encryption"
	aes256                     = "AES256"
)

// this is where is gets messy... :)
type PutOptions struct {
	ContentType          string
	ContentLength        int
	ContentEncoding      string
	AmzAcl               string
	ServerSideEncryption bool
	MetaHeader           http.Header
}

func (options *PutOptions) SetHeaders(req *http.Request) {
	headers := req.Header
	if options.MetaHeader != nil {
		for k := range options.MetaHeader {
			headers.Set("x-amz-meta-"+k, options.MetaHeader.Get(k))
		}
	}

	contentType := options.ContentType
	if contentType == "" {
		contentType = defaultContentType
	}
	headers.Set(headerContentType, contentType)

	if options.ContentEncoding != "" {
		headers.Add("Content-Encoding", options.ContentEncoding)
	}

	if options.AmzAcl != "" {
		headers.Add(headerAmzAcl, options.AmzAcl)
	}

	if options.ServerSideEncryption {
		headers.Add(headerServerSideEncruption, aes256)
	}
}

func (client *Client) putOld(bucket, key string, data []byte, options *PutOptions) error {
	if options == nil {
		options = &PutOptions{ContentType: defaultContentType}
	}

	buf := &bytes.Buffer{}
	csWriter := md5.New()
	r := io.TeeReader(bytes.NewReader(data), csWriter)
	if _, e := io.Copy(buf, r); e != nil {
		return e
	}

	m := csWriter.Sum(nil)
	var b []byte = m[0:]
	cs := b64.EncodeToString(b)
	req, e := http.NewRequest("PUT", client.keyUrl(bucket, key), buf)
	if e != nil {
		return e
	}

	req.Header.Set(headerContentMd5, cs)

	if options != nil {
		options.SetHeaders(req)
	}
	_, _, e = client.signSendAndRead(bucket, req)
	return e
}

func signPayload(payload, secret string, hashFunc func() hash.Hash) string {
	hash := hmac.New(hashFunc, []byte(secret))
	hash.Write([]byte(payload))
	signature := make([]byte, b64.EncodedLen(hash.Size()))
	b64.Encode(signature, hash.Sum(nil))
	return string(signature)
}

func s3Secret() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}

func s3Key() string {
	return os.Getenv("AWS_ACCESS_KEY_ID")
}

var s3ParamsToSign = map[string]struct{}{
	"acl":                          null,
	"location":                     null,
	"logging":                      null,
	"notification":                 null,
	"partNumber":                   null,
	"policy":                       null,
	"requestPayment":               null,
	"response-cache-control":       null,
	"response-content-disposition": null,
	"response-content-encoding":    null,
	"response-content-language":    null,
	"response-content-type":        null,
	"response-expires":             null,
	"torrent":                      null,
	"uploadId":                     null,
	"uploads":                      null,
	"versionId":                    null,
	"versioning":                   null,
	"versions":                     null,
}

func normalizeParams(url *url.URL) string {
	// instead of using url.Query() this does not change the encoding
	params := []string{}
	for _, part := range strings.Split(url.RawQuery, "&") {
		parts := strings.SplitN(part, "=", 2)
		if _, ok := s3ParamsToSign[parts[0]]; ok {
			params = append(params, part)
		}
	}
	sort.Strings(params)
	return strings.Join(params, "&")
}

func (client *Client) put(bucket, key string, in io.Reader, options *PutOptions) error {
	if options == nil {
		options = &PutOptions{ContentType: defaultContentType}
	}
	buf := &bytes.Buffer{}
	csWriter := md5.New()
	r := io.TeeReader(in, csWriter)
	_, e := io.CopyN(buf, r, minPartSize)
	if e == io.EOF {
		m := csWriter.Sum(nil)
		cs := b64.EncodeToString(m[0:])
		log.Printf("using %q", cs)
		req, e := http.NewRequest("PUT", client.keyUrl(bucket, key), buf)
		if e != nil {
			return e
		}
		options.SetHeaders(req)
		req.Header.Set(headerContentMd5, cs)
		_, _, e = client.signSendAndRead(bucket, req)
		return e
	} else if e == nil {
		mr := io.MultiReader(buf, r)
		mo := &MultipartOptions{
			PartSize: 5 * 1024 * 1024,
			Callback: func(res *UploadPartResult) {
				if res.Error != nil {
					log.Print("ERROR: " + e.Error())
				}
			},
			PutOptions: options,
		}
		_, e = putMultipart(bucket, key, mr, mo)
		return e
	} else {
		return e
	}
}

func initiateMultipartUpload(bucket, key string, opts *PutOptions) (result *InitiateMultipartUploadResult, e error) {
	theUrl := client.keyUrl(bucket, key) + "?uploads"
	req, e := http.NewRequest("POST", theUrl, nil)
	if e != nil {
		return nil, e
	}
	if opts != nil && opts.ServerSideEncryption {
		req.Header.Add(headerServerSideEncruption, aes256)
	}
	_, b, e := client.signSendAndRead(bucket, req)
	if e != nil {
		return nil, e
	}
	result = &InitiateMultipartUploadResult{}
	e = xml.Unmarshal(b, result)
	if e != nil {
		return nil, fmt.Errorf("ERROR: %s %s", e, string(b))
	}
	return result, e
}

func putMultipart(bucket, key string, f io.Reader, opts *MultipartOptions) (res *CompleteMultipartUploadResult, e error) {
	if opts == nil {
		opts = &MultipartOptions{PartSize: minPartSize}
	}
	if opts.PartSize == 0 {
		opts.PartSize = minPartSize
	}

	if opts.PartSize < minPartSize {
		return nil, fmt.Errorf("part size must be at least %d but was %d", minPartSize, opts.PartSize)
	}

	result, e := initiateMultipartUpload(bucket, key, opts.PutOptions)
	if e != nil {
		return nil, e
	}
	partId := 1
	parts := []*Part{}
	currentSize := int64(0)
	for {
		buf := bytes.NewBuffer(make([]byte, 0, opts.PartSize))
		i, e := io.CopyN(buf, f, int64(opts.PartSize))
		if e != nil && e != io.EOF {
			return nil, e
		}
		if i > 0 {
			part := &Part{PartNumber: partId, UploadId: result.UploadId, Bucket: bucket, Key: key, Reader: buf}
			e = part.execute(client)
			if opts.Callback != nil {
				opts.Callback(&UploadPartResult{Part: part, Error: e, CurrentSize: currentSize})
			}
			if e != nil {
				return nil, e
			}
			currentSize += i
			parts = append(parts, part)
			partId++
		}
		if e == io.EOF {
			break
		}
	}
	return completeMultipartUpload(bucket, key, result.UploadId, parts)
}

func completeMultipartUpload(bucket, key, uploadId string, parts []*Part) (result *CompleteMultipartUploadResult, e error) {
	theUrl := client.keyUrl(bucket, key) + fmt.Sprintf("?uploadId=%s", uploadId)
	payload := &CompleteMultipartUpload{Parts: parts}
	buf := &bytes.Buffer{}
	if e = xml.NewEncoder(buf).Encode(payload); e != nil {
		return nil, e
	}
	req, e := http.NewRequest("POST", theUrl, buf)
	if e != nil {
		return nil, e
	}
	_, b, e := client.signSendAndRead(bucket, req)
	if e != nil {
		return nil, e
	}
	result = &CompleteMultipartUploadResult{}
	return result, xml.Unmarshal(b, result)
}

func (m *Part) execute(client *Client) error {
	theUrl := client.keyUrl(m.Bucket, m.Key) + fmt.Sprintf("?partNumber=%d&uploadId=%s", m.PartNumber, m.UploadId)
	req, e := http.NewRequest("PUT", theUrl, m.Reader)
	if e != nil {
		return e
	}
	req.Header.Set("Content-Length", strconv.Itoa(m.ContentLength))
	rsp, _, e := client.signSendAndRead(m.Bucket, req)
	if e != nil {
		return e
	}
	m.ETag = rsp.Header.Get("ETag")
	return nil
}

type Client struct {
	Client *http.Client
	Key    string
	Secret string
	Region string
}

func (c *Client) signS3Request(req *http.Request, bucket string) {
	date := time.Now().UTC().Format(http.TimeFormat)
	payloadParts := []string{req.Method, req.Header.Get(headerContentMd5), req.Header.Get(headerContentType), date}
	amzHeaders := []string{}
	for k, v := range req.Header {
		value := strings.ToLower(k) + ":" + strings.Join(v, ",")
		if strings.HasPrefix(value, "x-amz") {
			amzHeaders = append(amzHeaders, value)
		}
	}
	sort.Strings(amzHeaders)
	payloadParts = append(payloadParts, amzHeaders...)
	path := req.URL.Path
	query := normalizeParams(req.URL)
	if query != "" {
		path += "?" + query
	}
	payloadParts = append(payloadParts, "/"+bucket+"/"+strings.TrimPrefix(path, "/"))
	payload := strings.Join(payloadParts, "\n")
	dbg.Printf("signing %q", payload)
	req.Header.Set(headerDate, date)
	req.Header.Set(headerAuthorization, "AWS "+s3Key()+":"+signPayload(payload, s3Secret(), sha1.New))
}

func (c *Client) keyUrl(bucket, key string) string {
	return "https://" + c.endpoint(bucket) + "/" + key
}

func (c *Client) endpoint(bucket string) string {
	return bucket + ".s3.amazonaws.com"
}

func (c *Client) signSendAndRead(bucket string, req *http.Request) (*http.Response, []byte, error) {
	req.Header.Set("Host", client.endpoint(bucket))
	c.signS3Request(req, bucket)
	rsp, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, nil, e
	}
	defer rsp.Body.Close()
	b, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return rsp, nil, e
	}
	if rsp.Status[0] != '2' {
		return rsp, nil, fmt.Errorf("expected status 2xx, got %s (%s)", rsp.Status, string(b))
	}
	return rsp, b, nil
}

type UploadPartResult struct {
	CurrentSize int64
	Part        *Part
	Error       error
}

type Part struct {
	PartNumber    int       `xml:"PartNumber"`
	ETag          string    `xml:"ETag"`
	UploadId      string    `xml:"-"`
	Bucket        string    `xml:"-"`
	Key           string    `xml:"-"`
	Reader        io.Reader `xml:"-"`
	ContentMD5    string    `xml:"-"`
	ContentLength int       `xml:"-"`
}

type MultipartOptions struct {
	*PutOptions
	PartSize int
	Callback func(*UploadPartResult)
}

type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string   `xml:"Location"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	ETag     string   `xml:"ETag"`
}

type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadId string   `xml:"UploadId"`
}

type CompleteMultipartUpload struct {
	XMLName xml.Name `xml:"CompleteMultipartUpload"`
	Parts   []*Part  `xml:"Part"`
}
