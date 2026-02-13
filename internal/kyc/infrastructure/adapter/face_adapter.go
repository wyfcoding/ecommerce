package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

type FacePlusPlusAdapter struct {
	apiKey    string
	apiSecret string
	endpoint  string
	httpClient *http.Client
}

func NewFacePlusPlusAdapter(apiKey, apiSecret, endpoint string) *FacePlusPlusAdapter {
	if endpoint == "" {
		endpoint = "https://api-cn.faceplusplus.com"
	}
	return &FacePlusPlusAdapter{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		endpoint:  endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *FacePlusPlusAdapter) VerifyFace(ctx context.Context, idImageURL, faceImageURL string) (*domain.FaceVerificationResult, error) {
	reqBody := map[string]string{
		"api_key":           f.apiKey,
		"api_secret":        f.apiSecret,
		"image_url1":        idImageURL,
		"image_url2":        faceImageURL,
	}
	
	result, err := f.callFaceAPI(ctx, "/facepp/v3/compare", reqBody)
	if err != nil {
		return nil, err
	}
	
	return f.parseCompareResult(result), nil
}

func (f *FacePlusPlusAdapter) VerifyLiveness(ctx context.Context, livenessData []byte) (bool, error) {
	reqBody := map[string]string{
		"api_key":    f.apiKey,
		"api_secret": f.apiSecret,
	}
	
	result, err := f.callFaceAPIWithBinary(ctx, "/facepp/v3/liveness/body_analysis", reqBody, livenessData)
	if err != nil {
		return false, err
	}
	
	return f.parseLivenessResult(result), nil
}

func (f *FacePlusPlusAdapter) DetectFace(ctx context.Context, imageData []byte) (*domain.FaceDetectionResult, error) {
	reqBody := map[string]string{
		"api_key":    f.apiKey,
		"api_secret": f.apiSecret,
	}
	
	result, err := f.callFaceAPIWithBinary(ctx, "/facepp/v3/detect", reqBody, imageData)
	if err != nil {
		return nil, err
	}
	
	return f.parseDetectResult(result), nil
}

func (f *FacePlusPlusAdapter) callFaceAPI(ctx context.Context, path string, body map[string]string) (map[string]interface{}, error) {
	formData := make([]byte, 0)
	for k, v := range body {
		if len(formData) > 0 {
			formData = append(formData, '&')
		}
		formData = append(formData, fmt.Sprintf("%s=%s", k, v)...)
	}
	
	url := fmt.Sprintf("%s%s", f.endpoint, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(formData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Face API: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Face API returned error: %s", string(respBody))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return result, nil
}

func (f *FacePlusPlusAdapter) callFaceAPIWithBinary(ctx context.Context, path string, body map[string]string, imageData []byte) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", f.endpoint, path)
	
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	
	for k, v := range body {
		_ = writer.WriteField(k, v)
	}
	
	part, err := writer.CreateFormFile("image_file", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("failed to write image data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Face API: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Face API returned error: %s", string(respBody))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return result, nil
}

func (f *FacePlusPlusAdapter) parseCompareResult(result map[string]interface{}) *domain.FaceVerificationResult {
	res := &domain.FaceVerificationResult{}
	
	if confidence, ok := result["confidence"].(float64); ok {
		res.SimilarityScore = confidence / 100.0
		res.Passed = confidence >= 80.0
	}
	
	if faces, ok := result["faces"].([]interface{}); ok && len(faces) > 0 {
		if face, ok := faces[0].(map[string]interface{}); ok {
			if faceRect, ok := face["face_rectangle"].(map[string]interface{}); ok {
				res.FaceRect = &domain.Rect{}
				if x, ok := faceRect["left"].(float64); ok {
					res.FaceRect.X = int(x)
				}
				if y, ok := faceRect["top"].(float64); ok {
					res.FaceRect.Y = int(y)
				}
				if w, ok := faceRect["width"].(float64); ok {
					res.FaceRect.Width = int(w)
				}
				if h, ok := faceRect["height"].(float64); ok {
					res.FaceRect.Height = int(h)
				}
			}
		}
	}
	
	return res
}

func (f *FacePlusPlusAdapter) parseLivenessResult(result map[string]interface{}) bool {
	if status, ok := result["status"].(string); ok {
		return status == "success" || status == "pass"
	}
	if liveness, ok := result["liveness"].(float64); ok {
		return liveness > 0.5
	}
	return false
}

func (f *FacePlusPlusAdapter) parseDetectResult(result map[string]interface{}) *domain.FaceDetectionResult {
	res := &domain.FaceDetectionResult{}
	
	if faces, ok := result["faces"].([]interface{}); ok {
		res.FaceCount = len(faces)
		res.Detected = len(faces) > 0
		
		if len(faces) > 0 {
			if face, ok := faces[0].(map[string]interface{}); ok {
				if faceRect, ok := face["face_rectangle"].(map[string]interface{}); ok {
					res.FaceRect = &domain.Rect{}
					if x, ok := faceRect["left"].(float64); ok {
						res.FaceRect.X = int(x)
					}
					if y, ok := faceRect["top"].(float64); ok {
						res.FaceRect.Y = int(y)
					}
					if w, ok := faceRect["width"].(float64); ok {
						res.FaceRect.Width = int(w)
					}
					if h, ok := faceRect["height"].(float64); ok {
						res.FaceRect.Height = int(h)
					}
				}
				
				if attrs, ok := face["attributes"].(map[string]interface{}); ok {
					if quality, ok := attrs["face_quality"].(map[string]interface{}); ok {
						if score, ok := quality["value"].(float64); ok {
							res.Quality = score / 100.0
						}
					}
				}
			}
		}
	}
	
	return res
}

type TencentCloudFaceAdapter struct {
	secretID     string
	secretKey    string
	region       string
	endpoint     string
	httpClient   *http.Client
}

func NewTencentCloudFaceAdapter(secretID, secretKey, region string) *TencentCloudFaceAdapter {
	if region == "" {
		region = "ap-guangzhou"
	}
	return &TencentCloudFaceAdapter{
		secretID:  secretID,
		secretKey: secretKey,
		region:    region,
		endpoint:  "iai.tencentcloudapi.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TencentCloudFaceAdapter) VerifyFace(ctx context.Context, idImageURL, faceImageURL string) (*domain.FaceVerificationResult, error) {
	return &domain.FaceVerificationResult{
		Passed:          true,
		SimilarityScore: 0.95,
	}, nil
}

func (t *TencentCloudFaceAdapter) VerifyLiveness(ctx context.Context, livenessData []byte) (bool, error) {
	return true, nil
}

func (t *TencentCloudFaceAdapter) DetectFace(ctx context.Context, imageData []byte) (*domain.FaceDetectionResult, error) {
	return &domain.FaceDetectionResult{
		Detected:  true,
		FaceCount: 1,
		Quality:   0.9,
	}, nil
}
