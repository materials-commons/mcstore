package mc

import "errors"

var (
	ErrInvalidProjectFilePath = errors.New("path not in project")
	ErrFileNotUploaded        = errors.New("existing file not uploaded")
	ErrFileVersionNotUploaded = errors.New("existing file with new version not uploaded")
)
