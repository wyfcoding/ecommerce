package adapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

type AliyunOCRAdapter struct {
	accessKeyID     string
	accessKeySecret string
	endpoint        string
	httpClient      *http.Client
}

func NewAliyunOCRAdapter(accessKeyID, accessKeySecret, endpoint string) *AliyunOCRAdapter {
	return &AliyunOCRAdapter{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		endpoint:        endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *AliyunOCRAdapter) RecognizeIDCard(ctx context.Context, imageData []byte, side domain.DocumentSide) (*domain.IDDocumentInfo, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	sideStr := "face"
	if side == domain.DocumentSideBack {
		sideStr = "back"
	}
	
	reqBody := map[string]interface{}{
		"image":     base64Image,
		"side":      sideStr,
		"configure": map[string]interface{}{
			"side": "face",
		},
	}
	
	result, err := a.callOCRAPI(ctx, "/ocr/idcard", reqBody)
	if err != nil {
		return nil, err
	}
	
	return a.parseIDCardResult(result, side), nil
}

func (a *AliyunOCRAdapter) RecognizePassport(ctx context.Context, imageData []byte) (*domain.IDDocumentInfo, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	reqBody := map[string]interface{}{
		"image": base64Image,
	}
	
	result, err := a.callOCRAPI(ctx, "/ocr/passport", reqBody)
	if err != nil {
		return nil, err
	}
	
	return a.parsePassportResult(result), nil
}

func (a *AliyunOCRAdapter) RecognizeDriversLicense(ctx context.Context, imageData []byte) (*domain.IDDocumentInfo, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	reqBody := map[string]interface{}{
		"image": base64Image,
	}
	
	result, err := a.callOCRAPI(ctx, "/ocr/driver-license", reqBody)
	if err != nil {
		return nil, err
	}
	
	return a.parseDriversLicenseResult(result), nil
}

func (a *AliyunOCRAdapter) RecognizeBusinessLicense(ctx context.Context, imageData []byte) (*domain.IDDocumentInfo, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	reqBody := map[string]interface{}{
		"image": base64Image,
	}
	
	result, err := a.callOCRAPI(ctx, "/ocr/business-license", reqBody)
	if err != nil {
		return nil, err
	}
	
	return a.parseBusinessLicenseResult(result), nil
}

func (a *AliyunOCRAdapter) callOCRAPI(ctx context.Context, path string, body interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	url := fmt.Sprintf("%s%s", a.endpoint, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("APPCODE %s", a.accessKeyID))
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OCR API: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCR API returned error: %s", string(respBody))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return result, nil
}

func (a *AliyunOCRAdapter) parseIDCardResult(result map[string]interface{}, side domain.DocumentSide) *domain.IDDocumentInfo {
	info := &domain.IDDocumentInfo{}
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return info
	}
	
	if side == domain.DocumentSideFront {
		if name, ok := data["name"].(string); ok {
			info.Name = name
		}
		if idNumber, ok := data["idNumber"].(string); ok {
			info.IDNumber = idNumber
		}
		if gender, ok := data["sex"].(string); ok {
			info.Gender = gender
		}
		if nationality, ok := data["nationality"].(string); ok {
			info.Nationality = nationality
		}
		if birthDate, ok := data["birthDate"].(string); ok {
			if t, err := time.Parse("20060102", birthDate); err == nil {
				info.BirthDate = &t
			}
		}
		if address, ok := data["address"].(string); ok {
			info.Address = address
		}
	} else {
		if issuingAuthority, ok := data["issue"].(string); ok {
			info.IssuingAuthority = issuingAuthority
		}
		if issueDate, ok := data["startDate"].(string); ok {
			if t, err := time.Parse("20060102", issueDate); err == nil {
				info.IssueDate = &t
			}
		}
		if expiryDate, ok := data["endDate"].(string); ok {
			if t, err := time.Parse("20060102", expiryDate); err == nil {
				info.ExpiryDate = &t
			}
		}
	}
	
	if confidence, ok := data["confidence"].(float64); ok {
		info.ConfidenceScore = confidence
	} else {
		info.ConfidenceScore = 0.95
	}
	
	return info
}

func (a *AliyunOCRAdapter) parsePassportResult(result map[string]interface{}) *domain.IDDocumentInfo {
	info := &domain.IDDocumentInfo{}
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return info
	}
	
	if name, ok := data["name"].(string); ok {
		info.Name = name
	}
	if idNumber, ok := data["idNumber"].(string); ok {
		info.IDNumber = idNumber
	}
	if gender, ok := data["sex"].(string); ok {
		info.Gender = gender
	}
	if nationality, ok := data["nationality"].(string); ok {
		info.Nationality = nationality
	}
	if birthDate, ok := data["birthDate"].(string); ok {
		if t, err := time.Parse("20060102", birthDate); err == nil {
			info.BirthDate = &t
		}
	}
	if expiryDate, ok := data["expiryDate"].(string); ok {
		if t, err := time.Parse("20060102", expiryDate); err == nil {
			info.ExpiryDate = &t
		}
	}
	
	info.ConfidenceScore = 0.90
	return info
}

func (a *AliyunOCRAdapter) parseDriversLicenseResult(result map[string]interface{}) *domain.IDDocumentInfo {
	info := &domain.IDDocumentInfo{}
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return info
	}
	
	if name, ok := data["name"].(string); ok {
		info.Name = name
	}
	if idNumber, ok := data["idNumber"].(string); ok {
		info.IDNumber = idNumber
	}
	if gender, ok := data["sex"].(string); ok {
		info.Gender = gender
	}
	if nationality, ok := data["nationality"].(string); ok {
		info.Nationality = nationality
	}
	if address, ok := data["address"].(string); ok {
		info.Address = address
	}
	if birthDate, ok := data["birthDate"].(string); ok {
		if t, err := time.Parse("20060102", birthDate); err == nil {
			info.BirthDate = &t
		}
	}
	
	info.ConfidenceScore = 0.90
	return info
}

func (a *AliyunOCRAdapter) parseBusinessLicenseResult(result map[string]interface{}) *domain.IDDocumentInfo {
	info := &domain.IDDocumentInfo{}
	
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return info
	}
	
	if name, ok := data["name"].(string); ok {
		info.Name = name
	}
	if regNumber, ok := data["regNumber"].(string); ok {
		info.IDNumber = regNumber
	}
	if address, ok := data["address"].(string); ok {
		info.Address = address
	}
	if issueDate, ok := data["establishDate"].(string); ok {
		if t, err := time.Parse("20060102", issueDate); err == nil {
			info.IssueDate = &t
		}
	}
	
	info.ConfidenceScore = 0.90
	return info
}
