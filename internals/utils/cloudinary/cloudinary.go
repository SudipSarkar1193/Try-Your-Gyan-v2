package cloudinary

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func Credentials() (*cloudinary.Cloudinary, context.Context, error) {
	// Initialize Cloudinary credentials and configuration
	cld, err := cloudinary.New()
	if err != nil {
		return nil, nil, err
	}
	cld.Config.URL.Secure = true // Ensures HTTPS URLs
	ctx := context.Background()
	return cld, ctx, nil
}

func UploadImage(cld *cloudinary.Cloudinary, ctx context.Context, base64Image string) (string, error) {
	// Generate a unique PublicID for the image
	uniquePublicID := "image_" + strconv.FormatInt(time.Now().Unix(), 10)

	// Upload the image
	resp, err := cld.Upload.Upload(ctx, base64Image, uploader.UploadParams{
		PublicID:       uniquePublicID,
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true),
	})

	// Handle errors during upload
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	// Return the secure URL of the uploaded image
	return resp.SecureURL, nil
}
