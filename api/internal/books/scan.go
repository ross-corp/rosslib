package books

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

// ScanBarcode accepts an uploaded image, detects an EAN-13 / ISBN barcode,
// looks up the ISBN via Open Library, upserts the book locally, and returns
// the normalised BookResult.
//
// POST /books/scan
// Content-Type: multipart/form-data
// Body: image file in "image" field
func (h *Handler) ScanBarcode(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not decode image — supported formats: JPEG, PNG, GIF"})
		return
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "could not process image for barcode detection"})
		return
	}

	reader := oned.NewEAN13Reader()
	result, err := reader.DecodeWithoutHints(bmp)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "no ISBN barcode detected in image",
			"hint":  "Make sure the barcode is clearly visible, well-lit, and not blurry.",
		})
		return
	}

	isbn := result.GetText()
	if isbn == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "barcode detected but could not read ISBN"})
		return
	}

	book, err := LookupBookByISBN(c.Request.Context(), h.pool, isbn, h.olClient, h.search)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("ISBN %s detected but book lookup failed", isbn)})
		return
	}
	if book == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("ISBN %s detected but no matching book found", isbn),
			"isbn":  isbn,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isbn": isbn,
		"book": book,
	})
}
