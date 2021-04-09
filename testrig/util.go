package testrig

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
)

// CreateMultipartFormData is a handy function for taking a fieldname and a filename, and creating a multipart form bytes buffer
// with the file contents set in the given fieldname. The extraFields param can be used to add extra FormFields to the request, as necessary.
// The returned bytes.Buffer b can be used like so:
// 	httptest.NewRequest(http.MethodPost, "https://example.org/whateverpath", bytes.NewReader(b.Bytes()))
// The returned *multipart.Writer w can be used to set the content type of the request, like so:
// 	req.Header.Set("Content-Type", w.FormDataContentType())
func CreateMultipartFormData(fieldName string, fileName string, extraFields map[string]string) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	var fw io.Writer
	file, err := os.Open(fileName)
	if err != nil {
		return b, nil, err
	}
	if fw, err = w.CreateFormFile(fieldName, file.Name()); err != nil {
		return b, nil, err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return b, nil, err
	}

	if extraFields != nil {
		for k, v := range extraFields {
			f, err := w.CreateFormField(k)
			if err != nil {
				return b, nil, err
			}
			if _, err := io.Copy(f, bytes.NewBufferString(v)); err != nil {
				return b, nil, err
			}
		}
	}

	if err := w.Close(); err != nil {
		return b, nil, err
	}
	return b, w, nil
}
