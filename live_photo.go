package onfido

import (
	"bytes"
	"context"
	"encoding/json"

	// "fmt"
	"io"
	"mime/multipart"

	// "net/http"
	// "net/textproto"
	//"os"
	// "strings"
	"time"
)

// type LivePhotoRequest struct {
//
// }

type LivePhoto struct {
	ID           string     `json:"id,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	Href         string     `json:"href,omitempty"`
	DownloadHref string     `json:"download_href,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	FileType     string     `json:"file_type,omitempty"`
	FileSize     int        `json:"file_size,omitempty"`
}

type LivePhotos struct {
	LivePhotos []*LivePhoto `json:"live_photos"`
}

func (c *Client) UploadLivePhoto(ctx context.Context, applicantID string, file io.ReadSeeker, filename string) (*LivePhoto, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := createFormFile(writer, "file", file, filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := writer.WriteField("applicant_id", applicantID); err != nil {
		return nil, err
	}
	// if err := writer.WriteField("side", string(dr.Side)); err != nil {
	// 	return nil, err
	// }
	// if err := writer.WriteField("issuing_country", string(dr.IssuingCountry)); err != nil {
	// 	return nil, err
	// }
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := c.newRequest("POST", "/live_photos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	var resp LivePhoto
	_, err = c.do(ctx, req, &resp)

	return &resp, err
}

func (c *Client) GetLivePhoto(ctx context.Context, id string) (*LivePhoto, error) {
	req, err := c.newRequest("GET", "/live_photos/"+id, nil)
	if err != nil {
		return nil, err
	}

	var resp LivePhoto
	_, err = c.do(ctx, req, &resp)
	return &resp, err
}

func (c *Client) GetLivePhotoDownload(ctx context.Context, id string) (*DocumentDownload, error) {
	req, err := c.newRequest("GET", "/live_photos/"+id+"/download", nil)
	if err != nil {
		return nil, err
	}

	var resp DocumentDownload
	bob, err = c.download(ctx, req, &resp)
	return &resp{Content: bob, Size: len(bob)}, err
}

// LivePhotoIter represents a LivePhoto iterator
type LivePhotoIter struct {
	*iter
}

// LivePhoto returns the current item in the iterator as a LivePhoto.
func (i *LivePhotoIter) LivePhoto() *LivePhoto {
	return i.Current().(*LivePhoto)
}

// ListLivePhotos retrieves the list of LivePhotos for the provided applicant.
// see https://LivePhotoation.onfido.com/?shell#list-LivePhotos
func (c *Client) ListLivePhotos(applicantID string) *LivePhotoIter {
	handler := func(body []byte) ([]interface{}, error) {
		var d LivePhotos
		if err := json.Unmarshal(body, &d); err != nil {
			return nil, err
		}

		values := make([]interface{}, len(d.LivePhotos))
		for i, v := range d.LivePhotos {
			values[i] = v
		}
		return values, nil
	}

	return &LivePhotoIter{&iter{
		c:       c,
		nextURL: "/live_photos?applicant_id=" + applicantID,
		handler: handler,
	}}
}
