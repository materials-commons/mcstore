package files

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db/schema"
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
	"application/matlab":                       "Matlab",
	"application/pdf":                          "PDF",
	"application/xml":                          "XML",
	"application/vnd.ms-excel":                 "MS-Excel",
	"image/bmp":                                "BMP",
	"image/x-ms-bmp":                           "BMP",
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
	if err := mime.AddExtensionType(".m", "application/matlab"); err != nil {
		app.Log.Errorf("AddExtensionType failed:", err)
	}

	var err error
	magic, err = magicmime.New(magicmime.MAGIC_MIME)
	if err != nil {
		app.Panicf("Unable to initialize magicmime: %s", err)
	}
}

// MediaType determines the mime media type for the given file. Because
// MaterialsCommons stores the file by id, which is different from the
// filename, the name and the path are passed. The name allows us to
// try and determine the file type by its extension.
func MediaType(name, path string) schema.MediaType {
	mtype := determineMediaType(name, path)
	m := schema.MediaType{
		Mime:        mtype,
		Description: getDescription(mtype),
	}
	return m
}

// determineMediaType determines the file mediatype first by checking
// its extension mime type, and then if that fails with libmagic. It
// returns "unknown" if the mediatype cannot be determined.
func determineMediaType(name, path string) string {
	mtype := mediaTypeByExtension(name)
	if mtype != "unknown" {
		return mtype
	}
	return mediaTypeByFile(path)
}

// mediaTypeByExtension determines the mediatype by the files extension.
// It returns "unknown" if this fails.
func mediaTypeByExtension(name string) string {
	ext := filepath.Ext(name)
	mtype := mime.TypeByExtension(ext)
	if mtype == "" {
		app.Log.Errorf("Unknown mediatype for extension: '%s'", ext)
		return "unknown"
	}
	return mtype
}

// mediaTypeByFile determines the mediatype by using libmagic. It returns
// "unknown" if libmagic cannot determine the type.
func mediaTypeByFile(path string) string {
	if !file.Exists(path) {
		app.Log.Errorf("Bad path for mediaTypeByFile: %s", path)
		return "unknown"
	}

	mtype, _ := magic.TypeByFile(path)
	if mtype == "" {
		app.Log.Errorf("Unknown magic mediatype for file: %s", path)
		return "unknown"
	}
	i := strings.Index(mtype, ";")
	if i == -1 {
		return mtype
	}

	return mtype[:i]
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
