package silverdile

import "github.com/disintegration/imaging"

type FormatType int

const (
	Default FormatType = iota
	PNG
	JPEG
)

func (ft FormatType) ImagingFormat() imaging.Format {
	switch ft {
	case PNG:
		return imaging.PNG
	case JPEG:
		return imaging.JPEG
	default:
		return imaging.PNG
	}
}

type ImageMeta struct {
	FormatType  FormatType
	ContentType string
}
