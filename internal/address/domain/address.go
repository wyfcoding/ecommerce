package domain

import (
	"errors"
	"time"
)

var (
	ErrAddressNotFound      = errors.New("address not found")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrAddressLimitExceeded = errors.New("address limit exceeded")
	ErrDefaultAddressExists = errors.New("default address already exists")
)

type AddressType string

const (
	AddressTypeHome   AddressType = "HOME"
	AddressTypeOffice AddressType = "OFFICE"
	AddressTypeSchool AddressType = "SCHOOL"
	AddressTypeOther  AddressType = "OTHER"
)

type AddressStatus int8

const (
	AddressStatusActive   AddressStatus = 1
	AddressStatusInactive AddressStatus = 2
	AddressStatusDeleted  AddressStatus = 3
)

type AddressVerificationStatus string

const (
	AddressVerificationUnverified AddressVerificationStatus = "UNVERIFIED"
	AddressVerificationPending    AddressVerificationStatus = "PENDING"
	AddressVerificationVerified   AddressVerificationStatus = "VERIFIED"
	AddressVerificationFailed     AddressVerificationStatus = "FAILED"
)

type UserAddress struct {
	ID                 uint                      `json:"id"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	UserID             uint64                    `json:"user_id"`
	Name               string                    `json:"name"`
	Phone              string                    `json:"phone"`
	CountryCode        string                    `json:"country_code"`
	Country            string                    `json:"country"`
	ProvinceCode       string                    `json:"province_code"`
	Province           string                    `json:"province"`
	CityCode           string                    `json:"city_code"`
	City               string                    `json:"city"`
	DistrictCode       string                    `json:"district_code"`
	District           string                    `json:"district"`
	Street             string                    `json:"street"`
	Detail             string                    `json:"detail"`
	PostalCode         string                    `json:"postal_code"`
	Longitude          float64                   `json:"longitude"`
	Latitude           float64                   `json:"latitude"`
	AddressType        AddressType               `json:"address_type"`
	Status             AddressStatus             `json:"status"`
	IsDefault          bool                      `json:"is_default"`
	VerificationStatus AddressVerificationStatus `json:"verification_status"`
	VerifiedAt         *time.Time                `json:"verified_at"`
	Tag                string                    `json:"tag"`
	Building           string                    `json:"building"`
	Floor              string                    `json:"floor"`
	Room               string                    `json:"room"`
	Landmark           string                    `json:"landmark"`
	Instructions       string                    `json:"instructions"`
	ContactName        string                    `json:"contact_name"`
	ContactPhone       string                    `json:"contact_phone"`
	UsageCount         int                       `json:"usage_count"`
	LastUsedAt         *time.Time                `json:"last_used_at"`
}

type AdministrativeDivision struct {
	ID         uint                      `json:"id"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
	Code       string                    `json:"code"`
	Name       string                    `json:"name"`
	ParentCode string                    `json:"parent_code"`
	Level      int                       `json:"level"`
	FullName   string                    `json:"full_name"`
	Pinyin     string                    `json:"pinyin"`
	ShortName  string                    `json:"short_name"`
	Enabled    bool                      `json:"enabled"`
	SortOrder  int                       `json:"sort_order"`
	Children   []*AdministrativeDivision `json:"children"`
}

type AddressValidation struct {
	ID                uint      `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	AddressID         uint      `json:"address_id"`
	OriginalAddress   string    `json:"original_address"`
	NormalizedAddress string    `json:"normalized_address"`
	IsValid           bool      `json:"is_valid"`
	Confidence        float64   `json:"confidence"`
	Suggestions       []string  `json:"suggestions"`
	ErrorMessage      string    `json:"error_message"`
	Provider          string    `json:"provider"`
}

type AddressSuggestion struct {
	Text       string  `json:"text"`
	Province   string  `json:"province"`
	City       string  `json:"city"`
	District   string  `json:"district"`
	Street     string  `json:"street"`
	PostalCode string  `json:"postal_code"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	Score      float64 `json:"score"`
}

type AddressParseResult struct {
	OriginalText string  `json:"original_text"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	Province     string  `json:"province"`
	City         string  `json:"city"`
	District     string  `json:"district"`
	Street       string  `json:"street"`
	Detail       string  `json:"detail"`
	PostalCode   string  `json:"postal_code"`
	Confidence   float64 `json:"confidence"`
}

func NewUserAddress(userID uint64, name, phone string) *UserAddress {
	return &UserAddress{
		UserID:             userID,
		Name:               name,
		Phone:              phone,
		CountryCode:        "CN",
		Country:            "中国",
		AddressType:        AddressTypeHome,
		Status:             AddressStatusActive,
		IsDefault:          false,
		VerificationStatus: AddressVerificationUnverified,
		UsageCount:         0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (a *UserAddress) SetLocation(province, provinceCode, city, cityCode, district, districtCode string) {
	a.Province = province
	a.ProvinceCode = provinceCode
	a.City = city
	a.CityCode = cityCode
	a.District = district
	a.DistrictCode = districtCode
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetDetail(street, detail, postalCode string) {
	a.Street = street
	a.Detail = detail
	a.PostalCode = postalCode
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetCoordinates(longitude, latitude float64) {
	a.Longitude = longitude
	a.Latitude = latitude
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetDefault(isDefault bool) {
	a.IsDefault = isDefault
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetType(addressType AddressType) {
	a.AddressType = addressType
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetTag(tag string) {
	a.Tag = tag
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetExtra(building, floor, room string) {
	a.Building = building
	a.Floor = floor
	a.Room = room
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetInstructions(instructions string) {
	a.Instructions = instructions
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) SetContact(name, phone string) {
	a.ContactName = name
	a.ContactPhone = phone
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) RecordUsage() {
	a.UsageCount++
	now := time.Now()
	a.LastUsedAt = &now
	a.UpdatedAt = now
}

func (a *UserAddress) Verify() {
	a.VerificationStatus = AddressVerificationVerified
	now := time.Now()
	a.VerifiedAt = &now
	a.UpdatedAt = now
}

func (a *UserAddress) FailVerification() {
	a.VerificationStatus = AddressVerificationFailed
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) Activate() {
	a.Status = AddressStatusActive
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) Deactivate() {
	a.Status = AddressStatusInactive
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) Delete() {
	a.Status = AddressStatusDeleted
	a.UpdatedAt = time.Now()
}

func (a *UserAddress) IsActive() bool {
	return a.Status == AddressStatusActive
}

func (a *UserAddress) IsVerified() bool {
	return a.VerificationStatus == AddressVerificationVerified
}

func (a *UserAddress) FullAddress() string {
	address := a.Province
	if a.City != a.Province {
		address += a.City
	}
	address += a.District + a.Street + a.Detail
	return address
}

func (a *UserAddress) FormatAddress() string {
	return a.Name + " " + a.Phone + " " + a.FullAddress()
}

func NewAdministrativeDivision(code, name, parentCode string, level int) *AdministrativeDivision {
	return &AdministrativeDivision{
		Code:       code,
		Name:       name,
		ParentCode: parentCode,
		Level:      level,
		FullName:   name,
		Enabled:    true,
		Children:   make([]*AdministrativeDivision, 0),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (d *AdministrativeDivision) AddChild(child *AdministrativeDivision) {
	d.Children = append(d.Children, child)
}

func (d *AdministrativeDivision) SetFullName(fullName string) {
	d.FullName = fullName
	d.UpdatedAt = time.Now()
}

func (d *AdministrativeDivision) SetPinyin(pinyin string) {
	d.Pinyin = pinyin
	d.UpdatedAt = time.Now()
}

func (d *AdministrativeDivision) Enable() {
	d.Enabled = true
	d.UpdatedAt = time.Now()
}

func (d *AdministrativeDivision) Disable() {
	d.Enabled = false
	d.UpdatedAt = time.Now()
}

func NewAddressValidation(addressID uint, originalAddress string) *AddressValidation {
	return &AddressValidation{
		AddressID:       addressID,
		OriginalAddress: originalAddress,
		IsValid:         false,
		Confidence:      0,
		Suggestions:     make([]string, 0),
		CreatedAt:       time.Now(),
	}
}

func (v *AddressValidation) SetResult(normalizedAddress string, isValid bool, confidence float64) {
	v.NormalizedAddress = normalizedAddress
	v.IsValid = isValid
	v.Confidence = confidence
}

func (v *AddressValidation) AddSuggestion(suggestion string) {
	v.Suggestions = append(v.Suggestions, suggestion)
}

func (v *AddressValidation) SetError(errMsg string) {
	v.ErrorMessage = errMsg
	v.IsValid = false
}

func (v *AddressValidation) SetProvider(provider string) {
	v.Provider = provider
}

type AddressRepository interface {
	FindByID(ctx any, id uint) (*UserAddress, error)
	FindByUserID(ctx any, userID uint64) ([]*UserAddress, error)
	FindActiveByUserID(ctx any, userID uint64) ([]*UserAddress, error)
	FindDefaultByUserID(ctx any, userID uint64) (*UserAddress, error)
	Save(ctx any, address *UserAddress) error
	Update(ctx any, address *UserAddress) error
	Delete(ctx any, id uint) error
	CountByUserID(ctx any, userID uint64) (int64, error)

	SaveDivision(ctx any, division *AdministrativeDivision) error
	FindDivisionByCode(ctx any, code string) (*AdministrativeDivision, error)
	FindDivisionsByParent(ctx any, parentCode string) ([]*AdministrativeDivision, error)
	FindDivisionsByLevel(ctx any, level int) ([]*AdministrativeDivision, error)
	SearchDivisions(ctx any, keyword string, limit int) ([]*AdministrativeDivision, error)

	SaveValidation(ctx any, validation *AddressValidation) error
	FindValidationByAddressID(ctx any, addressID uint) (*AddressValidation, error)
}

type AddressService interface {
	CreateAddress(ctx any, userID uint64, name, phone string, location *AddressLocation) (*UserAddress, error)
	UpdateAddress(ctx any, addressID uint, updates map[string]any) error
	DeleteAddress(ctx any, addressID uint) error
	SetDefaultAddress(ctx any, userID uint64, addressID uint) error
	GetUserAddresses(ctx any, userID uint64) ([]*UserAddress, error)
	ValidateAddress(ctx any, address *UserAddress) (*AddressValidation, error)
	ParseAddress(ctx any, text string) (*AddressParseResult, error)
	SuggestAddress(ctx any, keyword string, limit int) ([]*AddressSuggestion, error)
	NormalizeAddress(ctx any, address *UserAddress) (*UserAddress, error)
}

type AddressLocation struct {
	Province     string  `json:"province"`
	ProvinceCode string  `json:"province_code"`
	City         string  `json:"city"`
	CityCode     string  `json:"city_code"`
	District     string  `json:"district"`
	DistrictCode string  `json:"district_code"`
	Street       string  `json:"street"`
	Detail       string  `json:"detail"`
	PostalCode   string  `json:"postal_code"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
}
