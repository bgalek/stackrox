package zipdownload

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/rox/pkg/roxctl"
	"github.com/stackrox/rox/pkg/utils"
	"github.com/stackrox/rox/roxctl/common"
)

func writeZipToFolder(zipName, bundleType string) error {
	reader, err := zip.OpenReader(zipName)
	if err != nil {
		return err
	}

	outputFolder := strings.TrimRight(zipName, filepath.Ext(zipName))
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return errors.Wrapf(err, "Unable to create folder %q", outputFolder)
	}

	for _, f := range reader.File {
		fileReader, err := f.Open()
		if err != nil {
			return errors.Wrapf(err, "Unable to open file %q", f.Name)
		}
		data, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return errors.Wrapf(err, "Unable to read file %q", f.Name)
		}

		outputFile := filepath.Join(outputFolder, f.Name)
		folder := path.Dir(outputFile)
		if err := os.MkdirAll(folder, 0755); err != nil {
			return errors.Wrapf(err, "Unable to create folder %q", folder)
		}
		if err := ioutil.WriteFile(filepath.Join(outputFolder, f.Name), data, f.Mode()); err != nil {
			return errors.Wrapf(err, "Unable to write file %q", f.Name)
		}
	}
	printf("Successfully wrote %s folder %q\n", bundleType, outputFolder)
	return nil
}

func parseFilenameFromHeader(header http.Header) (string, error) {
	data := header.Get("Content-Disposition")
	if data == "" {
		return data, fmt.Errorf("could not parse filename from header: %+v", header)
	}
	data = strings.TrimPrefix(data, "attachment; filename=")
	return strings.Trim(data, `"`), nil
}

func printf(val string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, val, args...)
}

// GetZipOptions specifies a request to download a zip file
type GetZipOptions struct {
	Path, Method, BundleType string
	Body                     []byte
	Timeout                  time.Duration
	ExpandZip                bool
}

// GetZip downloads a zip from the given endpoint.
// bundleType is used for logging.
func GetZip(opts GetZipOptions) error {
	resp, err := common.DoHTTPRequestAndCheck200(opts.Path, opts.Timeout, opts.Method, bytes.NewBuffer(opts.Body))
	if err != nil {
		return err
	}
	defer utils.IgnoreError(resp.Body.Close)

	outputFilename, err := parseFilenameFromHeader(resp.Header)
	if err != nil {
		return err
	}
	// If containerized, then write a zip file
	if roxctl.InMainImage() {
		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			return errors.Wrap(err, "Error writing out zip file")
		}
		printf("Successfully wrote %s zip file\n", opts.BundleType)
	} else {
		file, err := os.Create(outputFilename)
		if err != nil {
			return errors.Wrapf(err, "Could not create file %q", outputFilename)
		}
		if _, err := io.Copy(file, resp.Body); err != nil {
			_ = file.Close()
			return errors.Wrap(err, "Error writing out zip file")
		}
		if err := file.Close(); err != nil {
			return errors.Wrap(err, "Error closing file")
		}
		if opts.ExpandZip {
			if err := writeZipToFolder(outputFilename, opts.BundleType); err != nil {
				return err
			}
		} else {
			printf("Successfully wrote %s folder %q\n", opts.BundleType, outputFilename)
		}
	}
	return nil
}
