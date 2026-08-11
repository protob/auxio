package imaging

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/davidbyttow/govips/v2/vips"
)

type Processor struct {
	maxWidth  int
	maxHeight int
}

func NewProcessor(maxWidth, maxHeight int) *Processor {
	return &Processor{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

func Startup() {
	vips.Startup(nil)
}

func Shutdown() {
	vips.Shutdown()
}

func (p *Processor) Transform(src io.Reader, params TransformParams) ([]byte, string, error) {
	srcBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source: %w", err)
	}

	img, err := vips.NewImageFromBuffer(srcBytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load image: %w", err)
	}
	defer img.Close()

	// Animated GIFs are flattened to a single frame - don't reject (upload still
	// works) but don't do it silently either.
	if img.Format() == vips.ImageTypeGIF && img.Pages() > 1 {
		slog.Warn("animated GIF flattened to a single frame during transform", "pages", img.Pages())
	}

	if params.Width > 0 || params.Height > 0 {
		if err := p.resize(img, params); err != nil {
			return nil, "", err
		}
	}

	return p.export(img, params)
}

func (p *Processor) resize(img *vips.ImageRef, params TransformParams) error {
	width := params.Width
	height := params.Height

	if width == 0 && height == 0 {
		return nil
	}

	imgWidth := img.Width()
	imgHeight := img.Height()

	var targetWidth int

	switch params.Fit {
	case "contain":
		if width > 0 && height > 0 {
			scaleW := float64(width) / float64(imgWidth)
			scaleH := float64(height) / float64(imgHeight)
			scale := scaleW
			if scaleH < scaleW {
				scale = scaleH
			}
			targetWidth = int(float64(imgWidth) * scale)
		} else if width > 0 {
			targetWidth = width
		} else {
			scale := float64(height) / float64(imgHeight)
			targetWidth = int(float64(imgWidth) * scale)
		}

	case "cover":
		if width > 0 && height > 0 {
			scaleW := float64(width) / float64(imgWidth)
			scaleH := float64(height) / float64(imgHeight)
			scale := scaleW
			if scaleH > scaleW {
				scale = scaleH
			}
			targetWidth = int(float64(imgWidth) * scale)
		} else if width > 0 {
			targetWidth = width
		} else {
			scale := float64(height) / float64(imgHeight)
			targetWidth = int(float64(imgWidth) * scale)
		}

	default:
		return fmt.Errorf("unknown fit mode: %s", params.Fit)
	}

	if err := img.Resize(float64(targetWidth)/float64(imgWidth), vips.KernelLanczos3); err != nil {
		return fmt.Errorf("failed to resize: %w", err)
	}

	if params.Fit == "cover" && width > 0 && height > 0 {
		// Clamp crop size to actual image dimensions (integer truncation during
		// resize can produce an image 1px smaller than the requested size).
		cropW := min(width, img.Width())
		cropH := min(height, img.Height())
		x := (img.Width() - cropW) / 2
		y := (img.Height() - cropH) / 2
		if err := img.ExtractArea(x, y, cropW, cropH); err != nil {
			return fmt.Errorf("failed to extract area: %w", err)
		}
	}

	return nil
}

func (p *Processor) export(img *vips.ImageRef, params TransformParams) ([]byte, string, error) {
	format := params.Format
	if format == "" {
		format = imageTypeToFormat(img.Format())
	}

	quality := params.Quality

	var buf []byte
	var contentType string
	var err error

	switch format {
	case "jpeg":
		buf, _, err = img.ExportJpeg(&vips.JpegExportParams{
			Quality: quality,
		})
		contentType = "image/jpeg"

	case "png":
		buf, _, err = img.ExportPng(&vips.PngExportParams{})
		contentType = "image/png"

	case "webp":
		buf, _, err = img.ExportWebp(&vips.WebpExportParams{
			Quality: quality,
		})
		contentType = "image/webp"

	case "avif":
		buf, _, err = img.ExportAvif(&vips.AvifExportParams{
			Quality: quality,
		})
		contentType = "image/avif"

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to export %s: %w", format, err)
	}

	return buf, contentType, nil
}

func imageTypeToFormat(t vips.ImageType) string {
	ext := t.FileExt()
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".webp":
		return "webp"
	case ".avif", ".heif":
		return "avif"
	case ".gif":
		return "gif"
	default:
		return "jpeg"
	}
}

func (p *Processor) MaxWidth() int {
	return p.maxWidth
}

func (p *Processor) MaxHeight() int {
	return p.maxHeight
}
