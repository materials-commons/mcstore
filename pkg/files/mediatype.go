package files

import (
	"mime"
	"path/filepath"

	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/rakyll/magicmime"
)

// maps media types to descriptions most people would recognize.
var mediaTypeDescriptions = map[string]string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "Spreadsheet",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   "Word",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "Presentation",
	"Composite Document File V2 Document, No summary info":                      "Composite Document File",
	"application/vnd.ms-powerpoint.presentation.macroEnabled.12":                "MS-PowerPoint",
	"text/xml":                                 "XML",
	"image/jpeg":                               "JPEG",
	"application/postscript":                   "Postscript",
	"image/png":                                "PNG",
	"application/json":                         "JSON",
	"image/vnd.ms-modi":                        "MS-Document Imaging",
	"application/vnd.ms-xpsdocument":           "MS-Postscript",
	"image/vnd.radiance":                       "Radiance",
	"application/vnd.sealedmedia.softseal.pdf": "Softseal PDF",
	"application/vnd.hp-PCL":                   "PCL",
	"application/xslt+xml":                     "XSLT",
	"image/gif":                                "GIF",
	"application/matlab":                       "matlab",
	"application/pdf":                          "PDF",
	"application/xml":                          "XML",
	"application/vnd.ms-excel":                 "MS-Excel",
	"image/bmp":                                "BMP",
	"image/tiff":                               "TIFF",
	"image/vnd.adobe.photoshop":                "Photoshop",
	"application/pkcs7-signature":              "PKCS",
	"image/vnd.dwg":                            "DWG",
	"application/octet-stream":                 "Binary",
	"application/rtf":                          "RTF",
	"text/plain":                               "Text",
	"application/vnd.ms-powerpoint":            "MS-PowerPoint",
	"application/x-troff-man":                  "TROFF",
	"video/x-ms-wmv":                           "WMV Video",
	"application/vnd.chemdraw+xml":             "ChemDraw",
	"text/html":                                "HTML",
	"video/mpeg":                               "MPEG Video",
	"text/csv":                                 "CSV",
	"application/zip":                          "ZIP",
	"application/msword":                       "MS-Word",
	"unknown":                                  "Unknown",
}

var magic *magicmime.Magic

func init() {
	mime.AddExtensionType(".m", "application/matlab")

	var err error
	magic, err = magicmime.New(magicmime.MAGIC_MIME)
	if err != nil {
		app.Panicf("Unable to initialize magicmime: %s", err)
	}
}

// MediaType determines the mime media type for the given file
func MediaType(path string) schema.MediaType {
	mtype := determineMediaType(path)
	return filloutMediaType(mtype)
}

// determineMediaType determines the file mediatype first by checking
// its extension mime type, and then if that fails with libmagic. It
// returns "unknown" if the mediatype cannot be determined.
func determineMediaType(path string) string {
	mtype := mediaTypeByExtension(path)
	if mtype != "unknown" {
		return mtype
	}
	return mediaTypeByFile(path)
}

// mediaTypeByExtension determines the mediatype by the files extension.
// It returns "unknown" if this fails.
func mediaTypeByExtension(path string) string {
	ext := "." + filepath.Ext(path)
	mtype := mime.TypeByExtension(ext)
	if mtype == "" {
		app.Log.Errorf("Unknown mediatype for extension: %s", ext)
		return "unknown"
	}
	return mtype
}

// mediaTypeByFile determines the mediatype by using libmagic. It returns
// "unknown" if libmagic cannot determine the type.
func mediaTypeByFile(path string) string {
	mtype, _ := magic.TypeByFile(path)
	if mtype == "" {
		app.Log.Errorf("Unknown magic mediatype for file: %s", path)
		return "unknown"
	}
	return mtype
}

// filloutMediaType creates a schema.MediaType and fills in its properties.
func filloutMediaType(mtype string) schema.MediaType {
	description := getDescription(mtype)
	m := schema.MediaType{
		Mime:        mtype,
		Description: description,
	}

	return m
}

// getDescription looks up the mediatype in the mediaTypeDescriptions map.
func getDescription(mtype string) string {
	description, found := mediaTypeDescriptions[mtype]
	if !found {
		app.Log.Debugf("mediatype '%s' not in description list", mtype)
		return "Unknown"
	}
	return description
}
