package handlers

import (
	"bytes"
	"context"
	_ "context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Milad-Abooali/4in-cs2skin-g1/src/internal/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Shared client with short timeout
var httpClient = &http.Client{Timeout: 2 * time.Second}

// md5(userID) as hex
func md5Hex(v interface{}) string {
	var s string
	switch val := v.(type) {
	case string:
		s = val
	case int64:
		s = strconv.FormatInt(val, 10)
	case int:
		s = strconv.Itoa(val)
	case float64: // اگر از gRPC NumberValue بیاد
		s = fmt.Sprintf("%.0f", val) // بدون اعشار
	default:
		s = fmt.Sprintf("%v", val) // fallback
	}
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// AvatarExists returns true if the static webp avatar exists.
func AvatarExists(userID int64) (string, bool) {
	hash := md5Hex(userID)
	url := "https://static.cs2skin.com/files/avatars/users/" + hash + ".webp"
	req, _ := http.NewRequest(http.MethodHead, url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return url, true
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// If your static server doesn't support HEAD and returns 405,
	// switch to a minimal GET (e.g., Range: bytes=0-0).
	if resp.StatusCode == http.StatusOK {
		return url, false
	}
	return url, true
}

// ClearAvatar Delete avatar from MinIO
func ClearAvatar(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// Get User ID
	userID, ok := GetUserId(email)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// MinIO Client
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	hash := md5Hex(userID)
	bucketName := "files"
	objectKey := "avatars/users/" + hash + ".webp"

	err = minioClient.RemoveObject(
		context.Background(),
		bucketName,
		objectKey,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		log.Fatalln(err)
	}

	resR.Type = "clearAvatar"
	resR.Data = ""
	return resR, errR
}

// extFromContentType returns a safe file extension for temp input.
func extFromContentType(ct string) string {
	ct = strings.ToLower(ct)
	switch {
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "webp"):
		return ".webp"
	default:
		return ".png"
	}
}

// ConvertToWebPBytes calls `cwebp` to convert any input to webp bytes.
func ConvertToWebPBytes(ctx context.Context, input []byte, contentType string, quality int) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	if strings.Contains(strings.ToLower(contentType), "webp") {
		return input, nil
	}

	cwebpPath, err := exec.LookPath("cwebp")
	if err != nil {
		return nil, fmt.Errorf("cwebp not found in PATH")
	}

	inExt := extFromContentType(contentType)
	inFile, err := os.CreateTemp("", "in-*"+inExt)
	if err != nil {
		return nil, fmt.Errorf("create temp input: %w", err)
	}
	defer os.Remove(inFile.Name())
	defer inFile.Close()
	if _, err = inFile.Write(input); err != nil {
		return nil, fmt.Errorf("write temp input: %w", err)
	}

	outFile, err := os.CreateTemp("", "out-*.webp")
	if err != nil {
		return nil, fmt.Errorf("create temp output: %w", err)
	}
	outName := outFile.Name()
	outFile.Close()
	defer os.Remove(outName)

	cmd := exec.CommandContext(ctx, cwebpPath, "-quiet", "-q", strconv.Itoa(quality), inFile.Name(), "-o", outName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cwebp failed: %v | %s", err, stderr.String())
	}

	outBytes, err := os.ReadFile(outName)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}
	return outBytes, nil
}

// UpdateAvatar Convert and upload to MinIO
func UpdateAvatar(data map[string]interface{}) (models.HandlerOK, models.HandlerError) {
	var (
		errR models.HandlerError
		resR models.HandlerOK
	)

	// Get Email
	email, ok := GetUserEmail(data)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	// Get User ID
	userID, ok := GetUserId(email)
	if !ok {
		errR.Type = "TOKEN_NOT_FOUND"
		errR.Code = 1032
		return resR, errR
	}

	if val, exists := data["avatar_base64"]; exists {
		if val == "" {
			errR.Type = "AVATAR_EMPTY"
			errR.Code = 1039
			return resR, errR
		}
	} else {
		errR.Type = "AVATAR_EXPECTED"
		errR.Code = 1037
		return resR, errR
	}
	rawB64, _ := data["avatar_base64"].(string)
	if rawB64 == "" {
		errR.Type = "AVATAR_EXPECTED"
		errR.Code = 1037
		return resR, errR
	}

	// remove data URL prefix if present
	if i := strings.Index(rawB64, ","); i != -1 && strings.HasPrefix(rawB64, "data:") {
		rawB64 = rawB64[i+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil || len(raw) == 0 {
		errR.Type = "INVALID_IMAGE_FORMAT"
		errR.Code = 1038
		return resR, errR
	}

	ct, _ := data["content_type"].(string)
	if ct == "" {
		if strings.HasPrefix(data["avatar_base64"].(string), "data:") {
			if j := strings.Index(rawB64, ";"); j != -1 {
				ct = rawB64[5:j]
			}
		}
	}

	// convert to webp via cwebp
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	webpBytes, err := ConvertToWebPBytes(ctx, raw, ct, 80)

	if err != nil {
		errR.Type = "INVALID_IMAGE_FORMAT"
		errR.Code = 1038
		return resR, errR
	}

	// MinIO Client
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	hash := md5Hex(userID)
	objectKey := "avatars/users/" + hash + ".webp"

	if err != nil {
		panic(err)
	}
	_, err = minioClient.PutObject(ctx, "files", objectKey, bytes.NewReader(webpBytes), int64(len(webpBytes)),
		minio.PutObjectOptions{
			ContentType:  "image/webp",
			CacheControl: "public, max-age=31536000",
		})
	if err != nil {
		errR.Type = "PROFILE_GRPC_ERROR"
		errR.Code = 1033
		return resR, errR
	}

	url := "https://static.cs2skin.com/files/" + objectKey

	resR.Type = "updateAvatar"
	resR.Data = url
	return resR, errR
}
