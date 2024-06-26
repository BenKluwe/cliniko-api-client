// Use of this source code is governed by the LGPL 2.1
// license that can be found in the LICENSE file.

package cliniko

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type UploadFileToS3BucketResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	XML201       *struct {
		PostResponse xml.Name `json:"postresponse" xml:"PostResponse"`
		Location     string   `json:"location" xml:"Location"`
		Bucket       string   `json:"bucket" xml:"Bucket"`
		Key          string   `json:"key" xml:"Key"`
		ETag         string   `json:"etag" xml:"ETag"`
	}
}

// ClinikoClientInterface is the interface specification
// for the client with extended functionality
type ClinikoClientInterface interface {
	CreateAttachment(
		ctx context.Context,
		patientId string,
		description *string,
		filename string,
		fileContent io.Reader,
		reqEditors ...RequestEditorFn,
	) (
		*PresignedPostGetResponse,
		*UploadFileToS3BucketResponse,
		*CreateUploadedPatientAttachmentPostResponse,
		error,
	)
}

// ClinikoClient builds on ClientWithResponsesInterface
// to provide extended functionality
type ClinikoClient struct {
	ClientWithResponsesInterface

	Client      *Client
	token       string
	vendor      string
	vendorEmail string
}

// NewClinikoClient creates a new Extended Client that wraps
// a ClientWithResponses type for advanced / additional functions
func NewClinikoClient(
	token string,
	vendor string,
	vendorEmail string,
	requestEditors ...RequestEditorFn,
) (
	*ClinikoClient, error,
) {
	var shard string
	tokenParts := strings.Split(token, "-")
	if len(tokenParts) == 1 {
		shard = "au1"
	} else {
		shard = tokenParts[1]
	}

	client, err := NewClient(
		fmt.Sprintf("https://api.%s.cliniko.com/v1", shard),
	)

	if err != nil {
		return nil, err
	}

	ret := &ClinikoClient{
		ClientWithResponsesInterface: &ClientWithResponses{client},
		Client:                       client,
		vendor:                       vendor,
		vendorEmail:                  vendorEmail,
		token: fmt.Sprintf(
			"Basic: %s",
			base64.StdEncoding.EncodeToString(
				[]byte(token+":"),
			)),
	}

	client.RequestEditors = append(client.RequestEditors, ret.addClinikoHeaders)
	client.RequestEditors = append(client.RequestEditors, requestEditors...)
	return ret, nil
}

func (c *ClinikoClient) addClinikoHeaders(
	ctx context.Context,
	req *http.Request,
) error {
	req.Header.Add("Authorization", c.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("%s (%s)", c.vendor, c.vendorEmail))
	return nil
}

// NewUploadFileToS3BucketPostRequest generates requests
// for UploadFileToS3Bucket
func (c *ClinikoClient) NewUploadFileToS3BucketPostRequest(
	presignedUrl *PresignedPostGetResponse,
	filename string,
	fileContent io.Reader,
) (
	*http.Request, error,
) {
	formFields := map[string]string{
		"acl":                   string(*presignedUrl.JSON200.Fields.Acl),
		"key":                   *presignedUrl.JSON200.Fields.Key,
		"policy":                *presignedUrl.JSON200.Fields.Policy,
		"success_action_status": string(*presignedUrl.JSON200.Fields.SuccessActionStatus),
		"x-amz-date":            time.Now().UTC().Format("20060102T150405Z"),
		"x-amz-algorithm":       *presignedUrl.JSON200.Fields.XAmzAlgorithm,
		"x-amz-credential":      *presignedUrl.JSON200.Fields.XAmzCredential,
		"x-amz-signature":       *presignedUrl.JSON200.Fields.XAmzSignature,
	}

	var form bytes.Buffer
	s3Form := multipart.NewWriter(&form)
	for key, val := range formFields {
		fw, err := s3Form.CreateFormField(key)
		if err != nil {
			return nil, err
		}
		fw.Write([]byte(val))
	}

	fw, err := s3Form.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, fileContent)
	if err != nil {
		return nil, err
	}
	s3Form.Close()

	req, err := http.NewRequest("POST", *presignedUrl.JSON200.Url, &form)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", s3Form.FormDataContentType())
	req.Header.Add("Accept", "application/xml")

	return req, nil
}

func (c *ClinikoClient) UploadFileToS3Bucket(
	ctx context.Context,
	presignedUrl *PresignedPostGetResponse,
	filename string,
	fileContent io.Reader,
) (
	*http.Response, error,
) {
	req, err :=
		c.NewUploadFileToS3BucketPostRequest(
			presignedUrl,
			filename,
			fileContent,
		)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	return c.Client.Client.Do(req)
}

// ParseUploadFileToS3BucketResponse parses an HTTP response
// from a CreateAttachment call
func (c *ClinikoClient) ParseUploadFileToS3BucketResponse(
	rsp *http.Response,
) (
	*UploadFileToS3BucketResponse, error,
) {

	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UploadFileToS3BucketResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "xml") && rsp.StatusCode == 201:
		var dest struct {
			PostResponse xml.Name `json:"postresponse" xml:"PostResponse"`
			Location     string   `json:"location" xml:"Location"`
			Bucket       string   `json:"bucket" xml:"Bucket"`
			Key          string   `json:"key" xml:"Key"`
			ETag         string   `json:"etag" xml:"ETag"`
		}
		if err := xml.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.XML201 = &dest
	}
	return response, nil
}

// CreateAttachment creates a given attachment using
// a filename and file content and attaches it to the
// given patient id. This function makes a total of 3 HTTP
// calls to 1. create a presigned Amazon S3 bucket URL
// 2. upload the file contents with the given name to
// the presigned url from 1.
// and 3. informs the Cliniko API of the new attachment
func (c *ClinikoClient) CreateAttachment(
	ctx context.Context,
	patientId string,
	description *string,
	filename string,
	fileContent io.Reader,
	reqEditors ...RequestEditorFn,
) (
	*PresignedPostGetResponse,
	*UploadFileToS3BucketResponse,
	*CreateUploadedPatientAttachmentPostResponse,
	error,
) {

	presignedUrl, err :=
		c.PresignedPostGetWithResponse(
			ctx,
			patientId,
			reqEditors...)

	if err != nil {
		return nil, nil, nil, err
	}

	if presignedUrl.JSON200 == nil {
		return presignedUrl, nil, nil, errors.New("presigned url request was unsuccessful")
	}

	rsp, err :=
		c.UploadFileToS3Bucket(
			ctx,
			presignedUrl,
			filename,
			fileContent)

	if err != nil {
		return presignedUrl, nil, nil, err
	}

	if fmt.Sprintf("%d", rsp.StatusCode) !=
		string(*presignedUrl.JSON200.Fields.SuccessActionStatus) {

		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, nil, nil, err
		}

		return presignedUrl, nil, nil,
			fmt.Errorf(
				"status code of s3 response not correct: %d, expected %s, response body: %s",
				rsp.StatusCode,
				*presignedUrl.JSON200.Fields.SuccessActionStatus,
				body,
			)
	}

	s3Response, err :=
		c.ParseUploadFileToS3BucketResponse(rsp)
	if err != nil {
		return presignedUrl, nil, nil, err
	}

	if s3Response.XML201 == nil {
		return presignedUrl, s3Response, nil, errors.New("s3 request was unsuccessful")
	}

	uploadUrl := fmt.Sprintf("%s/%s",
		*presignedUrl.JSON200.Url,
		s3Response.XML201.Key)

	attachmentPost := CreateUploadedPatientAttachmentPostJSONRequestBody{
		Description: description,
		PatientId:   &patientId,
		UploadUrl:   &uploadUrl,
	}

	attachmentPostResponse, err :=
		c.CreateUploadedPatientAttachmentPostWithResponse(
			ctx,
			attachmentPost,
			reqEditors...)

	if err != nil {
		return presignedUrl, s3Response, nil, err
	}

	if attachmentPostResponse.JSON201 == nil {
		return presignedUrl, s3Response, attachmentPostResponse, errors.New("post attachment to cliniko request was unsuccessful")
	}

	return presignedUrl, s3Response, attachmentPostResponse, nil
}
