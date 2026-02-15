// 生成摘要：
// - 从 internal/address 服务合并而来，作为 user 领域的子聚合
// - 地址属于用户领域的值对象/子实体，不应独立为微服务
// - 包含 UserAddress（用户收货地址）、AdministrativeDivision（行政区划）、
//   AddressValidation（地址校验结果）、AddressSuggestion（地址建议）等领域对象
// - 包含 UserAddressRepository 仓储接口

package domain

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrAddressNotFound 地址不存在
	ErrAddressNotFound = errors.New("address not found")
	// ErrInvalidAddress 无效地址
	ErrInvalidAddress = errors.New("invalid address")
	// ErrAddressLimitExceeded 地址数量超限
	ErrAddressLimitExceeded = errors.New("address limit exceeded")
)

// AddressType 地址类型
type AddressType string

const (
	AddressTypeHome   AddressType = "HOME"
	AddressTypeOffice AddressType = "OFFICE"
	AddressTypeSchool AddressType = "SCHOOL"
	AddressTypeOther  AddressType = "OTHER"
)

// AddressStatus 地址状态
type AddressStatus int8

const (
	AddressStatusActive   AddressStatus = 1
	AddressStatusInactive AddressStatus = 2
	AddressStatusDeleted  AddressStatus = 3
)

// AddressVerificationStatus 地址验证状态
type AddressVerificationStatus string

const (
	AddressVerificationUnverified AddressVerificationStatus = "UNVERIFIED"
	AddressVerificationPending    AddressVerificationStatus = "PENDING"
	AddressVerificationVerified   AddressVerificationStatus = "VERIFIED"
	AddressVerificationFailed     AddressVerificationStatus = "FAILED"
)

// UserAddress 用户收货地址实体
// 作为 User 聚合根的子实体，管理用户的收货地址信息
type UserAddress struct {
	ID                 uint                      `json:"id"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
	UserID             uint64                    `json:"user_id"`       // 所属用户 ID
	Name               string                    `json:"name"`          // 收件人姓名
	Phone              string                    `json:"phone"`         // 收件人手机号
	CountryCode        string                    `json:"country_code"`  // 国家编码
	Country            string                    `json:"country"`       // 国家名称
	ProvinceCode       string                    `json:"province_code"` // 省份编码
	Province           string                    `json:"province"`      // 省份名称
	CityCode           string                    `json:"city_code"`     // 城市编码
	City               string                    `json:"city"`          // 城市名称
	DistrictCode       string                    `json:"district_code"` // 区县编码
	District           string                    `json:"district"`      // 区县名称
	Street             string                    `json:"street"`        // 街道
	Detail             string                    `json:"detail"`        // 详细地址
	PostalCode         string                    `json:"postal_code"`   // 邮政编码
	Longitude          float64                   `json:"longitude"`     // 经度
	Latitude           float64                   `json:"latitude"`      // 纬度
	AddressType        AddressType               `json:"address_type"`  // 地址类型
	Status             AddressStatus             `json:"status"`        // 地址状态
	IsDefault          bool                      `json:"is_default"`    // 是否默认地址
	VerificationStatus AddressVerificationStatus `json:"verification_status"`
	VerifiedAt         *time.Time                `json:"verified_at"`
	Tag                string                    `json:"tag"`           // 标签
	Building           string                    `json:"building"`      // 楼栋
	Floor              string                    `json:"floor"`         // 楼层
	Room               string                    `json:"room"`          // 房间
	Landmark           string                    `json:"landmark"`      // 地标参照物
	Instructions       string                    `json:"instructions"`  // 配送说明
	ContactName        string                    `json:"contact_name"`  // 备用联系人
	ContactPhone       string                    `json:"contact_phone"` // 备用联系电话
	UsageCount         int                       `json:"usage_count"`   // 使用次数
	LastUsedAt         *time.Time                `json:"last_used_at"`  // 最后使用时间
}

// AdministrativeDivision 行政区划
// 管理国家→省→市→区多级行政区划数据
type AdministrativeDivision struct {
	ID         uint                      `json:"id"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
	Code       string                    `json:"code"`        // 行政区划编码
	Name       string                    `json:"name"`        // 名称
	ParentCode string                    `json:"parent_code"` // 上级编码
	Level      int                       `json:"level"`       // 层级（1=省 2=市 3=区）
	FullName   string                    `json:"full_name"`   // 完整名称
	Pinyin     string                    `json:"pinyin"`      // 拼音
	ShortName  string                    `json:"short_name"`  // 简称
	Enabled    bool                      `json:"enabled"`     // 是否启用
	SortOrder  int                       `json:"sort_order"`  // 排序
	Children   []*AdministrativeDivision `json:"children"`    // 子级区划
}

// AddressValidation 地址校验结果
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

// AddressSuggestion 地址搜索建议
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

// AddressParseResult 地址解析结果（从一段文字中自动识别地址信息）
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

// AddressLocation 地址定位信息
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

// NewUserAddress 创建新的用户收货地址
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

// SetLocation 设置地址位置（省市区）
func (a *UserAddress) SetLocation(province, provinceCode, city, cityCode, district, districtCode string) {
	a.Province = province
	a.ProvinceCode = provinceCode
	a.City = city
	a.CityCode = cityCode
	a.District = district
	a.DistrictCode = districtCode
	a.UpdatedAt = time.Now()
}

// SetDetail 设置详细地址
func (a *UserAddress) SetDetail(street, detail, postalCode string) {
	a.Street = street
	a.Detail = detail
	a.PostalCode = postalCode
	a.UpdatedAt = time.Now()
}

// SetCoordinates 设置经纬度
func (a *UserAddress) SetCoordinates(longitude, latitude float64) {
	a.Longitude = longitude
	a.Latitude = latitude
	a.UpdatedAt = time.Now()
}

// SetDefault 设置是否为默认地址
func (a *UserAddress) SetDefault(isDefault bool) {
	a.IsDefault = isDefault
	a.UpdatedAt = time.Now()
}

// RecordUsage 记录使用次数
func (a *UserAddress) RecordUsage() {
	a.UsageCount++
	now := time.Now()
	a.LastUsedAt = &now
	a.UpdatedAt = now
}

// Verify 标记为已验证
func (a *UserAddress) Verify() {
	a.VerificationStatus = AddressVerificationVerified
	now := time.Now()
	a.VerifiedAt = &now
	a.UpdatedAt = now
}

// ActivateAddress 激活地址
func (a *UserAddress) ActivateAddress() {
	a.Status = AddressStatusActive
	a.UpdatedAt = time.Now()
}

// DeleteAddress 删除地址（软删除）
func (a *UserAddress) DeleteAddress() {
	a.Status = AddressStatusDeleted
	a.UpdatedAt = time.Now()
}

// IsActive 是否为活跃地址
func (a *UserAddress) IsActive() bool {
	return a.Status == AddressStatusActive
}

// FullAddress 获取完整地址字符串
func (a *UserAddress) FullAddress() string {
	address := a.Province
	if a.City != a.Province {
		address += a.City
	}
	address += a.District + a.Street + a.Detail
	return address
}

// UserAddressRepository 用户地址仓储接口
type UserAddressRepository interface {
	// FindByID 根据 ID 查找地址
	FindByID(ctx context.Context, id uint) (*UserAddress, error)
	// FindByUserID 根据用户 ID 查找所有地址
	FindByUserID(ctx context.Context, userID uint64) ([]*UserAddress, error)
	// FindActiveByUserID 根据用户 ID 查找活跃地址
	FindActiveByUserID(ctx context.Context, userID uint64) ([]*UserAddress, error)
	// FindDefaultByUserID 查找用户默认地址
	FindDefaultByUserID(ctx context.Context, userID uint64) (*UserAddress, error)
	// Save 保存地址
	Save(ctx context.Context, address *UserAddress) error
	// Update 更新地址
	Update(ctx context.Context, address *UserAddress) error
	// Delete 删除地址
	Delete(ctx context.Context, id uint) error
	// CountByUserID 统计用户地址数量
	CountByUserID(ctx context.Context, userID uint64) (int64, error)

	// SaveDivision 保存行政区划
	SaveDivision(ctx context.Context, division *AdministrativeDivision) error
	// FindDivisionByCode 根据编码查找行政区划
	FindDivisionByCode(ctx context.Context, code string) (*AdministrativeDivision, error)
	// FindDivisionsByParent 根据上级编码查找子级区划
	FindDivisionsByParent(ctx context.Context, parentCode string) ([]*AdministrativeDivision, error)
	// SearchDivisions 模糊搜索行政区划
	SearchDivisions(ctx context.Context, keyword string, limit int) ([]*AdministrativeDivision, error)
}
