// Package drive reads the flyer and document folders through the Drive v3 REST API.
package drive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/fetch"
)

type File struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MimeType    string    `json:"mimeType"`
	CreatedTime time.Time `json:"createdTime"`
	MD5         string    `json:"md5Checksum"`
	Size        int64     `json:"size,string"`
}

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

const listFields = "nextPageToken,files(id,name,description,mimeType,createdTime,md5Checksum,size)"

// List returns the files directly inside folderID. Subfolders (like "Archief") are excluded on purpose:
// moving a flyer into one is how volunteers take it off the site.
func (c Client) List(ctx context.Context, folderID string) ([]File, error) {
	var all []File
	pageToken := ""
	for {
		q := url.Values{}
		q.Set("q", fmt.Sprintf("'%s' in parents and trashed = false and mimeType != 'application/vnd.google-apps.folder'", folderID))
		q.Set("fields", listFields)
		q.Set("pageSize", "100")
		q.Set("supportsAllDrives", "true")
		q.Set("includeItemsFromAllDrives", "true")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		var page struct {
			NextPageToken string `json:"nextPageToken"`
			Files         []File `json:"files"`
		}
		if err := fetch.JSON(ctx, c.HTTP, c.BaseURL+"/drive/v3/files?"+q.Encode(), &page); err != nil {
			return nil, fmt.Errorf("list drive folder: %w", err)
		}
		all = append(all, page.Files...)
		if page.NextPageToken == "" {
			return all, nil
		}
		pageToken = page.NextPageToken
	}
}

func (c Client) Download(ctx context.Context, fileID string, w io.Writer) error {
	resp, err := fetch.Get(ctx, c.HTTP, c.BaseURL+"/drive/v3/files/"+url.PathEscape(fileID)+"?alt=media&supportsAllDrives=true")
	if err != nil {
		return fmt.Errorf("download drive file %s: %w", fileID, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("download drive file %s: %w", fileID, err)
	}
	return nil
}
