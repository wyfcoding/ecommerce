package domain

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// WarehouseType 仓库类型
type WarehouseType string

const (
	WarehouseCentral     WarehouseType = "CENTRAL"     // 中心仓
	WarehouseRegional    WarehouseType = "REGIONAL"    // 区域仓
	WarehouseLocal       WarehouseType = "LOCAL"       // 本地仓
	WarehouseFulfillment WarehouseType = "FULFILLMENT" // 履约中心
)

// WarehouseStatus 仓库状态
type WarehouseStatus string

const (
	WarehouseActive      WarehouseStatus = "ACTIVE"      // 活跃
	WarehouseInactive    WarehouseStatus = "INACTIVE"    // 停用
	WarehouseMaintenance WarehouseStatus = "MAINTENANCE" // 维护中
)

// Warehouse 仓库
type Warehouse struct {
	ID            string          `json:"id"`
	WarehouseCode string          `json:"warehouse_code"`
	Name          string          `json:"name"`
	Type          WarehouseType   `json:"type"`
	Status        WarehouseStatus `json:"status"`

	// 位置信息
	Address   *Address `json:"address"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	TimeZone  string   `json:"time_zone"`

	// 容量信息
	TotalArea       float64 `json:"total_area"`       // 总面积（平方米）
	UsableArea      float64 `json:"usable_area"`      // 可用面积
	StorageCapacity float64 `json:"storage_capacity"` // 存储容量（立方米）
	CurrentUsage    float64 `json:"current_usage"`    // 当前使用率

	// 运营信息
	OpeningHours  *OpeningHours `json:"opening_hours"`
	ContactPerson string        `json:"contact_person"`
	ContactPhone  string        `json:"contact_phone"`
	ContactEmail  string        `json:"contact_email"`

	// 性能指标
	Throughput   float64 `json:"throughput"`    // 吞吐量（件/天）
	AccuracyRate float64 `json:"accuracy_rate"` // 准确率
	OnTimeRate   float64 `json:"on_time_rate"`  // 准时率
	DamageRate   float64 `json:"damage_rate"`   // 损坏率
	Priority     int     `json:"priority"`      // 优先级
	ShipCost     int64   `json:"ship_cost"`     // 基础配送成本

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastAudit time.Time `json:"last_audit"`
}

// Address 地址
type Address struct {
	Country    string `json:"country"`
	Province   string `json:"province"`
	City       string `json:"city"`
	District   string `json:"district"`
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
}

// OpeningHours 营业时间
type OpeningHours struct {
	Monday    *TimeRange  `json:"monday"`
	Tuesday   *TimeRange  `json:"tuesday"`
	Wednesday *TimeRange  `json:"wednesday"`
	Thursday  *TimeRange  `json:"thursday"`
	Friday    *TimeRange  `json:"friday"`
	Saturday  *TimeRange  `json:"saturday"`
	Sunday    *TimeRange  `json:"sunday"`
	Holidays  []time.Time `json:"holidays"`
}

// TimeRange 时间范围
type TimeRange struct {
	OpenTime  string `json:"open_time"`  // HH:MM
	CloseTime string `json:"close_time"` // HH:MM
}

// StorageLocation 存储位置
type StorageLocation struct {
	ID           string `json:"id"`
	WarehouseID  string `json:"warehouse_id"`
	LocationCode string `json:"location_code"`
	Zone         string `json:"zone"`  // 区域
	Aisle        string `json:"aisle"` // 通道
	Rack         string `json:"rack"`  // 货架
	Shelf        string `json:"shelf"` // 层
	Bin          string `json:"bin"`   // 货位

	// 容量信息
	MaxWeight     float64 `json:"max_weight"` // 最大承重（kg）
	MaxVolume     float64 `json:"max_volume"` // 最大体积（立方米）
	CurrentWeight float64 `json:"current_weight"`
	CurrentVolume float64 `json:"current_volume"`

	// 状态信息
	Status    string  `json:"status"` // AVAILABLE, OCCUPIED, RESERVED, MAINTENANCE
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`

	// 位置属性
	TemperatureZone string `json:"temperature_zone"` // 温区
	HumidityZone    string `json:"humidity_zone"`    // 湿度区
	SecurityLevel   string `json:"security_level"`   // 安全等级

	// 时间戳
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastInventory time.Time `json:"last_inventory"`
}

// WarehouseManager 仓库管理器
type WarehouseManager struct {
	warehouseRepo WarehouseRepository
	locationRepo  StorageLocationRepository
	inventoryRepo InventoryRepository
	mu            sync.RWMutex
	config        *WarehouseConfig
	warehouses    map[string]*Warehouse
}

// WarehouseConfig 仓库配置
type WarehouseConfig struct {
	DefaultZone    string  `json:"default_zone"`
	DefaultAisle   string  `json:"default_aisle"`
	MaxRacks       int     `json:"max_racks"`
	MaxShelves     int     `json:"max_shelves"`
	MaxBins        int     `json:"max_bins"`
	MinTemperature float64 `json:"min_temperature"`
	MaxTemperature float64 `json:"max_temperature"`
	MinHumidity    float64 `json:"min_humidity"`
	MaxHumidity    float64 `json:"max_humidity"`
}

// NewWarehouseManager 创建仓库管理器
func NewWarehouseManager(warehouseRepo WarehouseRepository, locationRepo StorageLocationRepository,
	inventoryRepo InventoryRepository) *WarehouseManager {

	return &WarehouseManager{
		warehouseRepo: warehouseRepo,
		locationRepo:  locationRepo,
		inventoryRepo: inventoryRepo,
		config: &WarehouseConfig{
			DefaultZone:    "A",
			DefaultAisle:   "01",
			MaxRacks:       100,
			MaxShelves:     10,
			MaxBins:        1000,
			MinTemperature: 15,
			MaxTemperature: 25,
			MinHumidity:    30,
			MaxHumidity:    70,
		},
		warehouses: make(map[string]*Warehouse),
	}
}

// Initialize 初始化仓库管理器
func (wm *WarehouseManager) Initialize(ctx context.Context) error {
	// 加载仓库信息
	warehouses, err := wm.warehouseRepo.GetActiveWarehouses(ctx)
	if err != nil {
		return fmt.Errorf("failed to load warehouses: %w", err)
	}

	wm.mu.Lock()
	for _, warehouse := range warehouses {
		wm.warehouses[warehouse.ID] = warehouse
	}
	wm.mu.Unlock()

	return nil
}

// CreateWarehouse 创建仓库
func (wm *WarehouseManager) CreateWarehouse(ctx context.Context, warehouse *Warehouse) error {
	// 验证仓库信息
	if err := wm.validateWarehouse(warehouse); err != nil {
		return fmt.Errorf("warehouse validation failed: %w", err)
	}

	// 生成仓库代码
	if warehouse.WarehouseCode == "" {
		warehouse.WarehouseCode = wm.generateWarehouseCode(warehouse)
	}

	// 设置默认值
	warehouse.Status = WarehouseActive
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	// 保存仓库
	err := wm.warehouseRepo.SaveWarehouse(ctx, warehouse)
	if err != nil {
		return fmt.Errorf("failed to save warehouse: %w", err)
	}

	// 更新缓存
	wm.mu.Lock()
	wm.warehouses[warehouse.ID] = warehouse
	wm.mu.Unlock()

	// 初始化存储位置
	go wm.initializeStorageLocations(ctx, warehouse)

	return nil
}

// validateWarehouse 验证仓库
func (wm *WarehouseManager) validateWarehouse(warehouse *Warehouse) error {
	if warehouse.Name == "" {
		return fmt.Errorf("warehouse name is required")
	}

	if warehouse.Address == nil {
		return fmt.Errorf("warehouse address is required")
	}

	if warehouse.Address.Country == "" {
		return fmt.Errorf("country is required")
	}

	if warehouse.Address.City == "" {
		return fmt.Errorf("city is required")
	}

	if warehouse.TotalArea <= 0 {
		return fmt.Errorf("total area must be positive")
	}

	if warehouse.StorageCapacity <= 0 {
		return fmt.Errorf("storage capacity must be positive")
	}

	return nil
}

// generateWarehouseCode 生成仓库代码
func (wm *WarehouseManager) generateWarehouseCode(warehouse *Warehouse) string {
	// 基于仓库类型和位置生成代码
	prefix := "WH"

	switch warehouse.Type {
	case WarehouseCentral:
		prefix = "WHC"
	case WarehouseRegional:
		prefix = "WHR"
	case WarehouseLocal:
		prefix = "WHL"
	case WarehouseFulfillment:
		prefix = "WHF"
	}

	cityCode := warehouse.Address.City[:3]
	timestamp := time.Now().UnixNano() % 10000

	return fmt.Sprintf("%s_%s_%04d", prefix, cityCode, timestamp)
}

// initializeStorageLocations 初始化存储位置
func (wm *WarehouseManager) initializeStorageLocations(ctx context.Context, warehouse *Warehouse) {
	// 根据仓库容量创建存储位置
	// 简化实现：创建固定数量的位置

	// 创建区域
	zones := []string{"A", "B", "C", "D"}

	for _, zone := range zones {
		// 创建通道
		for aisle := 1; aisle <= 10; aisle++ {
			// 创建货架
			for rack := 1; rack <= 10; rack++ {
				// 创建层
				for shelf := 1; shelf <= 5; shelf++ {
					// 创建货位
					for bin := 1; bin <= 10; bin++ {
						location := wm.createStorageLocation(warehouse, zone, aisle, rack, shelf, bin)

						err := wm.locationRepo.SaveStorageLocation(ctx, location)
						if err != nil {
							fmt.Printf("Failed to create storage location: %v\n", err)
							continue
						}
					}
				}
			}
		}
	}
}

// createStorageLocation 创建存储位置
func (wm *WarehouseManager) createStorageLocation(warehouse *Warehouse, zone string, aisle, rack, shelf, bin int) *StorageLocation {
	locationCode := fmt.Sprintf("%s-%02d-%03d-%02d-%02d", zone, aisle, rack, shelf, bin)

	return &StorageLocation{
		ID:              generateLocationID(),
		WarehouseID:     warehouse.ID,
		LocationCode:    locationCode,
		Zone:            zone,
		Aisle:           fmt.Sprintf("%02d", aisle),
		Rack:            fmt.Sprintf("%03d", rack),
		Shelf:           fmt.Sprintf("%02d", shelf),
		Bin:             fmt.Sprintf("%02d", bin),
		MaxWeight:       1000, // 1吨
		MaxVolume:       10,   // 10立方米
		Status:          "AVAILABLE",
		TemperatureZone: "NORMAL",
		HumidityZone:    "NORMAL",
		SecurityLevel:   "STANDARD",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// FindOptimalLocation 查找最优存储位置
func (wm *WarehouseManager) FindOptimalLocation(ctx context.Context, warehouseID, productID string,
	quantity float64, weight, volume float64, requirements *StorageRequirements) (*StorageLocation, error) {

	// 获取可用位置
	availableLocations, err := wm.locationRepo.GetAvailableLocations(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available locations: %w", err)
	}

	// 过滤符合要求的位置
	var candidateLocations []*StorageLocation
	for _, location := range availableLocations {
		if wm.isLocationSuitable(location, weight, volume, requirements) {
			candidateLocations = append(candidateLocations, location)
		}
	}

	if len(candidateLocations) == 0 {
		return nil, fmt.Errorf("no suitable location found")
	}

	// 选择最优位置
	optimalLocation := wm.selectOptimalLocation(candidateLocations, productID, quantity)

	return optimalLocation, nil
}

// isLocationSuitable 检查位置是否合适
func (wm *WarehouseManager) isLocationSuitable(location *StorageLocation, weight, volume float64,
	requirements *StorageRequirements) bool {

	// 检查承重能力
	if weight > location.MaxWeight-location.CurrentWeight {
		return false
	}

	// 检查体积容量
	if volume > location.MaxVolume-location.CurrentVolume {
		return false
	}

	// 检查温区要求
	if requirements != nil && requirements.TemperatureZone != "" {
		if location.TemperatureZone != requirements.TemperatureZone {
			return false
		}
	}

	// 检查湿度要求
	if requirements != nil && requirements.HumidityZone != "" {
		if location.HumidityZone != requirements.HumidityZone {
			return false
		}
	}

	// 检查安全等级
	if requirements != nil && requirements.SecurityLevel != "" {
		if location.SecurityLevel != requirements.SecurityLevel {
			return false
		}
	}

	return true
}

// selectOptimalLocation 选择最优位置
func (wm *WarehouseManager) selectOptimalLocation(locations []*StorageLocation, productID string, quantity float64) *StorageLocation {
	if len(locations) == 0 {
		return nil
	}

	// 使用评分系统选择最优位置
	var bestLocation *StorageLocation
	var bestScore float64 = -1

	for _, location := range locations {
		score := wm.calculateLocationScore(location, productID, quantity)

		if score > bestScore {
			bestScore = score
			bestLocation = location
		}
	}

	return bestLocation
}

// calculateLocationScore 计算位置评分
func (wm *WarehouseManager) calculateLocationScore(location *StorageLocation, productID string, quantity float64) float64 {
	var score float64

	// 1. 位置利用率（越低越好）
	weightUtilization := location.CurrentWeight / location.MaxWeight
	volumeUtilization := location.CurrentVolume / location.MaxVolume
	utilizationScore := 1 - (weightUtilization+volumeUtilization)/2
	score += utilizationScore * 0.4

	// 2. 位置便利性（基于区域和通道）
	convenienceScore := wm.calculateConvenienceScore(location)
	score += convenienceScore * 0.3

	// 3. 产品相似性（如果已有相同产品）
	if location.ProductID == productID {
		score += 0.2
	}

	// 4. 最近盘点时间（越近越好）
	daysSinceInventory := time.Since(location.LastInventory).Hours() / 24
	inventoryScore := math.Max(0, 1-daysSinceInventory/30) // 30天内为满分
	score += inventoryScore * 0.1

	return score
}

// calculateConvenienceScore 计算便利性评分
func (wm *WarehouseManager) calculateConvenienceScore(location *StorageLocation) float64 {
	// 简化实现：基于区域和通道计算
	// A区、01通道的位置最方便

	var score float64

	// 区域评分
	switch location.Zone {
	case "A":
		score += 1.0
	case "B":
		score += 0.8
	case "C":
		score += 0.6
	case "D":
		score += 0.4
	default:
		score += 0.5
	}

	// 通道评分（数字越小越方便）
	aisleNum := 0
	fmt.Sscanf(location.Aisle, "%d", &aisleNum)
	if aisleNum > 0 {
		aisleScore := math.Max(0, 1-float64(aisleNum-1)/20)
		score += aisleScore
	}

	return score / 2 // 归一化到0-1
}

// AllocateLocation 分配存储位置
func (wm *WarehouseManager) AllocateLocation(ctx context.Context, warehouseID, productID string,
	quantity float64, weight, volume float64) (*StorageLocation, error) {

	// 查找最优位置
	location, err := wm.FindOptimalLocation(ctx, warehouseID, productID, quantity, weight, volume, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find optimal location: %w", err)
	}

	// 更新位置状态
	location.Status = "OCCUPIED"
	location.ProductID = productID
	location.Quantity = quantity
	location.CurrentWeight += weight
	location.CurrentVolume += volume
	location.UpdatedAt = time.Now()

	// 保存更新
	err = wm.locationRepo.UpdateStorageLocation(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to update storage location: %w", err)
	}

	return location, nil
}

// ReleaseLocation 释放存储位置
func (wm *WarehouseManager) ReleaseLocation(ctx context.Context, locationID string, quantity, weight, volume float64) error {
	// 获取位置
	location, err := wm.locationRepo.GetStorageLocation(ctx, locationID)
	if err != nil {
		return fmt.Errorf("failed to get storage location: %w", err)
	}

	// 更新位置状态
	location.Quantity -= quantity
	location.CurrentWeight -= weight
	location.CurrentVolume -= volume

	if location.Quantity <= 0 {
		location.Status = "AVAILABLE"
		location.ProductID = ""
		location.Quantity = 0
	}

	location.UpdatedAt = time.Now()

	// 保存更新
	err = wm.locationRepo.UpdateStorageLocation(ctx, location)
	if err != nil {
		return fmt.Errorf("failed to update storage location: %w", err)
	}

	return nil
}

// GetWarehouseInventory 获取仓库库存
func (wm *WarehouseManager) GetWarehouseInventory(ctx context.Context, warehouseID string) (*WarehouseInventory, error) {
	// 获取仓库信息
	warehouse, err := wm.warehouseRepo.GetWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	// 获取库存数据
	inventoryItems, err := wm.inventoryRepo.GetInventoryByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	// 计算库存统计
	stats := wm.calculateInventoryStats(inventoryItems)

	// 计算仓库使用率
	usageRate := wm.calculateWarehouseUsage(warehouse, stats)

	inventory := &WarehouseInventory{
		Warehouse:      warehouse,
		InventoryItems: inventoryItems,
		Statistics:     stats,
		UsageRate:      usageRate,
		GeneratedAt:    time.Now(),
	}

	return inventory, nil
}

// calculateInventoryStats 计算库存统计
func (wm *WarehouseManager) calculateInventoryStats(items []*InventoryItem) *InventoryStats {
	stats := &InventoryStats{
		TotalSKUs:         0,
		TotalQuantity:     0,
		TotalValue:        0,
		CategoryBreakdown: make(map[string]float64),
		ValueBreakdown:    make(map[string]float64),
	}

	for _, item := range items {
		if item.ProductID != "" {
			stats.TotalSKUs++
		}

		stats.TotalQuantity += float64(item.Quantity)

		// 假设 InventoryItem 中包含成本信息
		// 如果没有，需要从 Product 仓库获取
		value := float64(item.Quantity) * item.UnitCost
		stats.TotalValue += value

		// 按类别统计
		if item.Category != "" {
			stats.CategoryBreakdown[item.Category] += float64(item.Quantity)
			stats.ValueBreakdown[item.Category] += value
		}
	}

	return stats
}

// calculateWarehouseUsage 计算仓库使用率
func (wm *WarehouseManager) calculateWarehouseUsage(warehouse *Warehouse, stats *InventoryStats) *WarehouseUsage {
	usage := &WarehouseUsage{
		WarehouseID:   warehouse.ID,
		AreaUsage:     0,
		VolumeUsage:   0,
		CapacityUsage: 0,
		SKUDensity:    0,
		ValueDensity:  0,
	}

	// 计算面积使用率
	if warehouse.TotalArea > 0 {
		usage.AreaUsage = stats.TotalQuantity * 0.1 / warehouse.TotalArea // 假设每件商品占用0.1平方米
	}

	// 计算体积使用率
	if warehouse.StorageCapacity > 0 {
		usage.VolumeUsage = stats.TotalQuantity * 0.01 / warehouse.StorageCapacity // 假设每件商品占用0.01立方米
	}

	// 计算容量使用率
	usage.CapacityUsage = (usage.AreaUsage + usage.VolumeUsage) / 2

	// 计算SKU密度
	if warehouse.UsableArea > 0 {
		usage.SKUDensity = float64(stats.TotalSKUs) / warehouse.UsableArea
	}

	// 计算价值密度
	if warehouse.UsableArea > 0 {
		usage.ValueDensity = stats.TotalValue / warehouse.UsableArea
	}

	return usage
}

// Data structures

type StorageRequirements struct {
	TemperatureZone string     `json:"temperature_zone"`
	HumidityZone    string     `json:"humidity_zone"`
	SecurityLevel   string     `json:"security_level"`
	SpecialHandling bool       `json:"special_handling"`
	Hazardous       bool       `json:"hazardous"`
	ExpiryDate      *time.Time `json:"expiry_date"`
}

type WarehouseInventory struct {
	Warehouse      *Warehouse       `json:"warehouse"`
	InventoryItems []*InventoryItem `json:"inventory_items"`
	Statistics     *InventoryStats  `json:"statistics"`
	UsageRate      *WarehouseUsage  `json:"usage_rate"`
	GeneratedAt    time.Time        `json:"generated_at"`
}

type InventoryStats struct {
	TotalSKUs         int                `json:"total_skus"`
	TotalQuantity     float64            `json:"total_quantity"`
	TotalValue        float64            `json:"total_value"`
	CategoryBreakdown map[string]float64 `json:"category_breakdown"`
	ValueBreakdown    map[string]float64 `json:"value_breakdown"`
}

type WarehouseUsage struct {
	WarehouseID   string  `json:"warehouse_id"`
	AreaUsage     float64 `json:"area_usage"`
	VolumeUsage   float64 `json:"volume_usage"`
	CapacityUsage float64 `json:"capacity_usage"`
	SKUDensity    float64 `json:"sku_density"`
	ValueDensity  float64 `json:"value_density"`
}

// Repository interfaces

type StorageLocationRepository interface {
	SaveStorageLocation(ctx context.Context, location *StorageLocation) error
	GetStorageLocation(ctx context.Context, locationID string) (*StorageLocation, error)
	GetStorageLocationByCode(ctx context.Context, warehouseID, locationCode string) (*StorageLocation, error)
	GetAvailableLocations(ctx context.Context, warehouseID string) ([]*StorageLocation, error)
	GetLocationsByProduct(ctx context.Context, warehouseID, productID string) ([]*StorageLocation, error)
	UpdateStorageLocation(ctx context.Context, location *StorageLocation) error
	DeleteStorageLocation(ctx context.Context, locationID string) error
}

// Helper functions

func generateLocationID() string {
	return fmt.Sprintf("LOC_%d", time.Now().UnixNano())
}
