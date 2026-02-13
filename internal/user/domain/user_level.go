package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidUserLevel      = errors.New("invalid user level")
	ErrInsufficientPrivilege = errors.New("insufficient privilege")
	ErrBenefitNotAvailable   = errors.New("benefit not available for current level")
	ErrLevelExpired          = errors.New("user level has expired")
)

type UserLevel string

const (
	UserLevelNormal    UserLevel = "NORMAL"
	UserLevelBronze    UserLevel = "BRONZE"
	UserLevelSilver    UserLevel = "SILVER"
	UserLevelGold      UserLevel = "GOLD"
	UserLevelPlatinum  UserLevel = "PLATINUM"
	UserLevelDiamond   UserLevel = "DIAMOND"
	UserLevelBlackGold UserLevel = "BLACK_GOLD"
)

func (l UserLevel) Order() int {
	switch l {
	case UserLevelNormal:
		return 0
	case UserLevelBronze:
		return 1
	case UserLevelSilver:
		return 2
	case UserLevelGold:
		return 3
	case UserLevelPlatinum:
		return 4
	case UserLevelDiamond:
		return 5
	case UserLevelBlackGold:
		return 6
	default:
		return 0
	}
}

func (l UserLevel) IsHigherOrEqual(other UserLevel) bool {
	return l.Order() >= other.Order()
}

func (l UserLevel) NextLevel() UserLevel {
	switch l {
	case UserLevelNormal:
		return UserLevelBronze
	case UserLevelBronze:
		return UserLevelSilver
	case UserLevelSilver:
		return UserLevelGold
	case UserLevelGold:
		return UserLevelPlatinum
	case UserLevelPlatinum:
		return UserLevelDiamond
	case UserLevelDiamond:
		return UserLevelBlackGold
	default:
		return l
	}
}

func (l UserLevel) PrevLevel() UserLevel {
	switch l {
	case UserLevelBronze:
		return UserLevelNormal
	case UserLevelSilver:
		return UserLevelBronze
	case UserLevelGold:
		return UserLevelSilver
	case UserLevelPlatinum:
		return UserLevelGold
	case UserLevelDiamond:
		return UserLevelPlatinum
	case UserLevelBlackGold:
		return UserLevelDiamond
	default:
		return l
	}
}

type BenefitType string

const (
	BenefitTypeDiscount      BenefitType = "DISCOUNT"
	BenefitTypeFreeShipping  BenefitType = "FREE_SHIPPING"
	BenefitTypePriority      BenefitType = "PRIORITY"
	BenefitTypeExclusive     BenefitType = "EXCLUSIVE"
	BenefitTypePoints        BenefitType = "POINTS_BONUS"
	BenefitTypeCoupon        BenefitType = "COUPON"
	BenefitTypeService       BenefitType = "SERVICE"
	BenefitTypeReturn        BenefitType = "RETURN"
	BenefitTypeBirthday      BenefitType = "BIRTHDAY"
	BenefitTypeEarlyAccess   BenefitType = "EARLY_ACCESS"
)

type UserLevelInfo struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UserID          uint64    `json:"user_id"`
	CurrentLevel    UserLevel `json:"current_level"`
	CurrentPoints   int64     `json:"current_points"`
	TotalPoints     int64     `json:"total_points"`
	LevelStartDate  *time.Time `json:"level_start_date"`
	LevelExpiryDate *time.Time `json:"level_expiry_date"`
	NextLevelPoints int64     `json:"next_level_points"`
	IsLifetime      bool      `json:"is_lifetime"`
	UpgradeHistory  []*LevelUpgradeRecord `json:"upgrade_history"`
}

type LevelUpgradeRecord struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       uint64    `json:"user_id"`
	FromLevel    UserLevel `json:"from_level"`
	ToLevel      UserLevel `json:"to_level"`
	PointsAtTime int64     `json:"points_at_time"`
	Reason       string    `json:"reason"`
	UpgradeAt    time.Time `json:"upgrade_at"`
}

type LevelBenefit struct {
	ID           uint        `json:"id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Level        UserLevel   `json:"level"`
	BenefitType  BenefitType `json:"benefit_type"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Value        float64     `json:"value"`
	Unit         string      `json:"unit"`
	MaxUsage     int         `json:"max_usage"`
	Period       string      `json:"period"`
	Icon         string      `json:"icon"`
	Enabled      bool        `json:"enabled"`
	Priority     int         `json:"priority"`
}

type LevelConfig struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Level           UserLevel `json:"level"`
	Name            string    `json:"name"`
	MinPoints       int64     `json:"min_points"`
	MaxPoints       int64     `json:"max_points"`
	Discount        float64   `json:"discount"`
	PointsRate      float64   `json:"points_rate"`
	FreeShipping    bool      `json:"free_shipping"`
	PrioritySupport bool      `json:"priority_support"`
	ExclusiveItems  bool      `json:"exclusive_items"`
	ReturnDays      int       `json:"return_days"`
	BirthdayBonus   float64   `json:"birthday_bonus"`
	Color           string    `json:"color"`
	Icon            string    `json:"icon"`
	Description     string    `json:"description"`
	Enabled         bool      `json:"enabled"`
}

type UserBenefitUsage struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UserID     uint64    `json:"user_id"`
	BenefitID  uint      `json:"benefit_id"`
	UsageCount int       `json:"usage_count"`
	Period     string    `json:"period"`
	ResetAt    *time.Time `json:"reset_at"`
}

func NewUserLevelInfo(userID uint64) *UserLevelInfo {
	now := time.Now()
	return &UserLevelInfo{
		UserID:         userID,
		CurrentLevel:   UserLevelNormal,
		CurrentPoints:  0,
		TotalPoints:    0,
		IsLifetime:     false,
		UpgradeHistory: make([]*LevelUpgradeRecord, 0),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (l *UserLevelInfo) AddPoints(points int64, configs []LevelConfig) bool {
	l.CurrentPoints += points
	l.TotalPoints += points
	l.UpdatedAt = time.Now()
	return l.checkUpgrade(configs)
}

func (l *UserLevelInfo) DeductPoints(points int64) error {
	if l.CurrentPoints < points {
		return errors.New("insufficient points")
	}
	l.CurrentPoints -= points
	l.UpdatedAt = time.Now()
	return nil
}

func (l *UserLevelInfo) checkUpgrade(configs []LevelConfig) bool {
	for _, config := range configs {
		if !config.Enabled {
			continue
		}
		if l.CurrentPoints >= config.MinPoints && l.CurrentLevel.Order() < config.Level.Order() {
			l.upgrade(config.Level, "points_upgrade")
			return true
		}
	}
	return false
}

func (l *UserLevelInfo) upgrade(newLevel UserLevel, reason string) {
	oldLevel := l.CurrentLevel
	l.CurrentLevel = newLevel
	now := time.Now()
	l.LevelStartDate = &now
	
	expiryDate := now.AddDate(1, 0, 0)
	l.LevelExpiryDate = &expiryDate
	
	l.UpdatedAt = now
	
	record := &LevelUpgradeRecord{
		UserID:       l.UserID,
		FromLevel:    oldLevel,
		ToLevel:      newLevel,
		PointsAtTime: l.CurrentPoints,
		Reason:       reason,
		UpgradeAt:    now,
		CreatedAt:    now,
	}
	l.UpgradeHistory = append(l.UpgradeHistory, record)
}

func (l *UserLevelInfo) Downgrade(reason string) {
	if l.CurrentLevel == UserLevelNormal {
		return
	}
	oldLevel := l.CurrentLevel
	l.CurrentLevel = l.CurrentLevel.PrevLevel()
	now := time.Now()
	l.LevelStartDate = &now
	l.UpdatedAt = now
	
	record := &LevelUpgradeRecord{
		UserID:       l.UserID,
		FromLevel:    oldLevel,
		ToLevel:      l.CurrentLevel,
		PointsAtTime: l.CurrentPoints,
		Reason:       reason,
		UpgradeAt:    now,
		CreatedAt:    now,
	}
	l.UpgradeHistory = append(l.UpgradeHistory, record)
}

func (l *UserLevelInfo) SetLifetime() {
	l.IsLifetime = true
	l.LevelExpiryDate = nil
	l.UpdatedAt = time.Now()
}

func (l *UserLevelInfo) IsExpired() bool {
	if l.IsLifetime || l.LevelExpiryDate == nil {
		return false
	}
	return time.Now().After(*l.LevelExpiryDate)
}

func (l *UserLevelInfo) Renew(duration time.Duration) {
	expiry := time.Now().Add(duration)
	l.LevelExpiryDate = &expiry
	l.UpdatedAt = time.Now()
}

func (l *UserLevelInfo) GetProgress(configs []LevelConfig) float64 {
	nextLevel := l.CurrentLevel.NextLevel()
	if nextLevel == l.CurrentLevel {
		return 100.0
	}
	
	for _, config := range configs {
		if config.Level == nextLevel {
			if config.MinPoints <= 0 {
				return 100.0
			}
			progress := float64(l.CurrentPoints) / float64(config.MinPoints) * 100
			if progress > 100 {
				progress = 100
			}
			return progress
		}
	}
	return 0
}

func (l *UserLevelInfo) GetNextLevelPoints(configs []LevelConfig) int64 {
	nextLevel := l.CurrentLevel.NextLevel()
	if nextLevel == l.CurrentLevel {
		return 0
	}
	
	for _, config := range configs {
		if config.Level == nextLevel {
			return config.MinPoints - l.CurrentPoints
		}
	}
	return 0
}

func NewLevelConfig(level UserLevel, name string, minPoints, maxPoints int64) *LevelConfig {
	return &LevelConfig{
		Level:      level,
		Name:       name,
		MinPoints:  minPoints,
		MaxPoints:  maxPoints,
		Discount:   1.0,
		PointsRate: 1.0,
		ReturnDays: 7,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (c *LevelConfig) CanUpgrade(currentPoints int64) bool {
	return currentPoints >= c.MinPoints && currentPoints <= c.MaxPoints
}

func NewLevelBenefit(level UserLevel, benefitType BenefitType, name string, value float64) *LevelBenefit {
	return &LevelBenefit{
		Level:       level,
		BenefitType: benefitType,
		Name:        name,
		Value:       value,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

type LevelRepository interface {
	FindLevelInfoByUserID(ctx interface{}, userID uint64) (*UserLevelInfo, error)
	SaveLevelInfo(ctx interface{}, info *UserLevelInfo) error
	UpdateLevelInfo(ctx interface{}, info *UserLevelInfo) error
	
	FindLevelConfigs(ctx interface{}) ([]LevelConfig, error)
	FindLevelConfigByLevel(ctx interface{}, level UserLevel) (*LevelConfig, error)
	SaveLevelConfig(ctx interface{}, config *LevelConfig) error
	
	FindBenefitsByLevel(ctx interface{}, level UserLevel) ([]LevelBenefit, error)
	FindAllBenefits(ctx interface{}) ([]LevelBenefit, error)
	SaveBenefit(ctx interface{}, benefit *LevelBenefit) error
	
	FindUpgradeHistory(ctx interface{}, userID uint64) ([]LevelUpgradeRecord, error)
	SaveUpgradeRecord(ctx interface{}, record *LevelUpgradeRecord) error
	
	FindBenefitUsage(ctx interface{}, userID uint64, benefitID uint) (*UserBenefitUsage, error)
	SaveBenefitUsage(ctx interface{}, usage *UserBenefitUsage) error
}

type LevelService interface {
	GetUserLevel(userID uint64) (*UserLevelInfo, error)
	AddPoints(userID uint64, points int64, reason string) error
	GetAvailableBenefits(userID uint64) ([]LevelBenefit, error)
	UseBenefit(userID uint64, benefitID uint) error
	CheckPrivilege(userID uint64, requiredLevel UserLevel) bool
	CalculatePoints(orderAmount int64) int64
}
