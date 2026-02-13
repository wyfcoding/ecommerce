package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

type RiskAssessmentAdapter struct {
	blacklistProvider  BlacklistProvider
	sanctionsProvider  SanctionsProvider
}

type BlacklistProvider interface {
	Check(ctx context.Context, idNumber string) (bool, string, error)
}

type SanctionsProvider interface {
	Check(ctx context.Context, name string) (bool, error)
}

func NewRiskAssessmentAdapter(blacklist BlacklistProvider, sanctions SanctionsProvider) *RiskAssessmentAdapter {
	return &RiskAssessmentAdapter{
		blacklistProvider:  blacklist,
		sanctionsProvider:  sanctions,
	}
}

func (r *RiskAssessmentAdapter) CalculateRiskScore(ctx context.Context, app *domain.KYCApplication) (*domain.RiskAssessmentResult, error) {
	result := &domain.RiskAssessmentResult{
		Score:   0,
		Level:   "LOW",
		Factors: []domain.RiskFactor{},
	}
	
	factors := r.calculateRiskFactors(ctx, app)
	totalScore := 0
	totalWeight := 0.0
	
	for _, factor := range factors {
		totalScore += int(float64(factor.Score) * factor.Weight)
		totalWeight += factor.Weight
	}
	
	if totalWeight > 0 {
		result.Score = int(float64(totalScore) / totalWeight)
	}
	
	result.Factors = factors
	result.Level = r.scoreToLevel(result.Score)
	
	return result, nil
}

func (r *RiskAssessmentAdapter) CheckBlacklist(ctx context.Context, idNumber string) (bool, string, error) {
	if r.blacklistProvider == nil {
		return false, "", nil
	}
	return r.blacklistProvider.Check(ctx, idNumber)
}

func (r *RiskAssessmentAdapter) CheckSanctionsList(ctx context.Context, name string) (bool, error) {
	if r.sanctionsProvider == nil {
		return false, nil
	}
	return r.sanctionsProvider.Check(ctx, name)
}

func (r *RiskAssessmentAdapter) calculateRiskFactors(ctx context.Context, app *domain.KYCApplication) []domain.RiskFactor {
	factors := []domain.RiskFactor{}
	
	factors = append(factors, r.checkIDAge(app))
	factors = append(factors, r.checkIDValidity(app))
	factors = append(factors, r.checkFaceSimilarity(app))
	factors = append(factors, r.checkOCRConfidence(app))
	factors = append(factors, r.checkDocumentQuality(app))
	
	return factors
}

func (r *RiskAssessmentAdapter) checkIDAge(app *domain.KYCApplication) domain.RiskFactor {
	factor := domain.RiskFactor{
		FactorType:  "ID_AGE",
		Description: "ID document age risk",
		Score:       0,
		Weight:      0.15,
	}
	
	if app.BirthDate != nil {
		age := calculateAge(*app.BirthDate)
		if age < 18 {
			factor.Score = 80
			factor.Description = "Underage applicant"
		} else if age < 25 {
			factor.Score = 20
			factor.Description = "Young applicant"
		} else if age > 70 {
			factor.Score = 30
			factor.Description = "Elderly applicant"
		} else {
			factor.Score = 0
			factor.Description = "Normal age range"
		}
	}
	
	return factor
}

func (r *RiskAssessmentAdapter) checkIDValidity(app *domain.KYCApplication) domain.RiskFactor {
	factor := domain.RiskFactor{
		FactorType:  "ID_VALIDITY",
		Description: "ID document validity",
		Score:       0,
		Weight:      0.20,
	}
	
	for _, doc := range app.Documents {
		if doc.OCRInfo != nil && doc.OCRInfo.ExpiryDate != nil {
			if doc.OCRInfo.ExpiryDate.Before(app.CreatedAt) {
				factor.Score = 100
				factor.Description = "ID document has expired"
				return factor
			}
			
			daysUntilExpiry := int(doc.OCRInfo.ExpiryDate.Sub(app.CreatedAt).Hours() / 24)
			if daysUntilExpiry < 30 {
				factor.Score = 50
				factor.Description = "ID document expiring soon"
				return factor
			}
		}
	}
	
	factor.Description = "ID document valid"
	return factor
}

func (r *RiskAssessmentAdapter) checkFaceSimilarity(app *domain.KYCApplication) domain.RiskFactor {
	factor := domain.RiskFactor{
		FactorType:  "FACE_SIMILARITY",
		Description: "Face verification similarity",
		Score:       0,
		Weight:      0.25,
	}
	
	if app.FaceVerification != nil {
		if !app.FaceVerification.Passed {
			factor.Score = 100
			factor.Description = "Face verification failed"
		} else if app.FaceVerification.SimilarityScore < 0.8 {
			factor.Score = 60
			factor.Description = fmt.Sprintf("Low face similarity: %.2f", app.FaceVerification.SimilarityScore)
		} else if app.FaceVerification.SimilarityScore < 0.9 {
			factor.Score = 20
			factor.Description = fmt.Sprintf("Medium face similarity: %.2f", app.FaceVerification.SimilarityScore)
		} else {
			factor.Score = 0
			factor.Description = fmt.Sprintf("High face similarity: %.2f", app.FaceVerification.SimilarityScore)
		}
		
		if !app.FaceVerification.LivenessPassed {
			factor.Score = 100
			factor.Description = "Liveness check failed"
		}
	}
	
	return factor
}

func (r *RiskAssessmentAdapter) checkOCRConfidence(app *domain.KYCApplication) domain.RiskFactor {
	factor := domain.RiskFactor{
		FactorType:  "OCR_CONFIDENCE",
		Description: "OCR recognition confidence",
		Score:       0,
		Weight:      0.20,
	}
	
	totalConfidence := 0.0
	docCount := 0
	
	for _, doc := range app.Documents {
		if doc.OCRInfo != nil && doc.OCRInfo.ConfidenceScore > 0 {
			totalConfidence += doc.OCRInfo.ConfidenceScore
			docCount++
		}
	}
	
	if docCount > 0 {
		avgConfidence := totalConfidence / float64(docCount)
		if avgConfidence < 0.7 {
			factor.Score = 80
			factor.Description = fmt.Sprintf("Low OCR confidence: %.2f", avgConfidence)
		} else if avgConfidence < 0.85 {
			factor.Score = 30
			factor.Description = fmt.Sprintf("Medium OCR confidence: %.2f", avgConfidence)
		} else {
			factor.Score = 0
			factor.Description = fmt.Sprintf("High OCR confidence: %.2f", avgConfidence)
		}
	}
	
	return factor
}

func (r *RiskAssessmentAdapter) checkDocumentQuality(app *domain.KYCApplication) domain.RiskFactor {
	factor := domain.RiskFactor{
		FactorType:  "DOCUMENT_QUALITY",
		Description: "Document completeness",
		Score:       0,
		Weight:      0.20,
	}
	
	requiredDocs := app.GetRequiredDocuments()
	uploadedTypes := make(map[domain.IDType]bool)
	
	for _, doc := range app.Documents {
		uploadedTypes[doc.DocumentType] = true
	}
	
	missingCount := 0
	for _, requiredType := range requiredDocs {
		if !uploadedTypes[requiredType] {
			missingCount++
		}
	}
	
	if missingCount > 0 {
		factor.Score = 50 * missingCount
		factor.Description = fmt.Sprintf("Missing %d required documents", missingCount)
	} else {
		factor.Description = "All required documents uploaded"
	}
	
	return factor
}

func (r *RiskAssessmentAdapter) scoreToLevel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 30:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}
	return age
}

type MockBlacklistProvider struct{}

func NewMockBlacklistProvider() *MockBlacklistProvider {
	return &MockBlacklistProvider{}
}

func (m *MockBlacklistProvider) Check(ctx context.Context, idNumber string) (bool, string, error) {
	return false, "", nil
}

type MockSanctionsProvider struct{}

func NewMockSanctionsProvider() *MockSanctionsProvider {
	return &MockSanctionsProvider{}
}

func (m *MockSanctionsProvider) Check(ctx context.Context, name string) (bool, error) {
	return false, nil
}
