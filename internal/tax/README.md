# 税务服务 (Tax Service) 功能说明

## 新增功能概览

本次实现为电商系统新增了完整的税务服务功能，包括增值税、消费税、关税计算引擎，跨境税务规则配置，第三方税务服务集成，以及税务报表生成。

## 1. 税务计算引擎

### 1.1 增值税(VAT)计算引擎
位置：`internal/tax/domain/calculators.go`

支持三种计算方法：
- **标准增值税** (VATMethodStandard): 价外税，税额 = 不含税金额 × 税率
- **内含增值税** (VATMethodInclusive): 倒算不含税金额，不含税金额 = 含税金额 / (1 + 税率)
- **级联增值税** (VATMethodCascading): 在已有的税额基础上再计税

### 1.2 消费税(Excise)计算引擎
支持三种计税基础：
- **从价计征** (ExciseBaseAdValorem): 按商品价值的一定比例征收
- **从量计征** (ExciseBaseSpecific): 按商品数量征收固定税额
- **复合计征** (ExciseBaseCompound): 同时采用从价和从量两种方式

### 1.3 关税(Duty)计算引擎
支持四种计算类型：
- **从价税** (DutyTypeAdValorem): 按完税价格的一定比例征收
- **从量税** (DutyTypeSpecific): 按商品数量征收固定税额
- **复合税** (DutyTypeCompound): 同时采用从价和从量两种方式
- **选择税** (DutyTypeAlternative): 选择从价或从量中较高的税额

### 1.4 综合税务计算器
统一计算多种税费，自动汇总增值税、消费税、关税。

## 2. 跨境税务规则配置

位置：`internal/tax/domain/crossborder.go`

### 2.1 核心功能
- **跨境税务配置**: 支持配置不同国家/地区之间的税务规则
- **贸易类型支持**: B2B, B2C, C2C, 一般贸易进出口, 过境贸易
- **交易模式**: DDP, DDU, DAP, EXW, FOB, CIF等
- **税款征收方式**: IOSS, OSS, 海关征收, 卖家代扣代缴, 买家自行申报

### 2.2 优惠贸易协定
- 支持配置自由贸易协定(FTA)
- 原产地规则验证(ROO): 完全获得、区域价值成分、税则归类改变等
- 优惠税率自动应用

### 2.3 最低征税门槛(De Minimis)
- 低于门槛值的交易免税
- 各国门槛值可配置

### 2.4 欧盟特殊支持
- **OSS配置**: 一站式申报系统配置
- **IOSS配置**: 进口一站式申报系统配置
- **OSS申报**: 按季度申报欧盟内B2C销售

## 3. 第三方税务服务集成

位置：`internal/tax/infrastructure/thirdparty/`

### 3.1 支持的提供商

#### Avalara ( avalara.go )
- 全球税务计算和合规服务
- 地址验证和标准化
- 税率查询
- 交易记录管理

#### Vertex ( vertex.go )
- 企业级税务解决方案
- 复杂的税务规则引擎
- 税收区域查找

### 3.2 核心特性
- **统一接口**: ThirdPartyTaxProvider 接口统一不同提供商的API
- **降级机制**: FallbackProvider 在第三方服务不可用时自动切换到本地计算
- **健康检查**: 实时监控第三方服务可用性
- **配置管理**: 支持多环境配置

### 3.3 使用方式
```go
// 配置Avalara
config := &thirdparty.ProviderConfig{
    ProviderName: "avalara",
    APIBaseURL:   "https://rest.avatax.com",
    APIKey:       "your-api-key",
    CompanyCode:  "your-company-code",
    Timeout:      30 * time.Second,
}

// 创建提供商
factory := thirdparty.NewProviderFactory()
factory.RegisterConfig("avalara", config)
provider, _ := factory.CreateProvider("avalara")

// 计算税费
request := &thirdparty.TaxCalculationRequest{
    TransactionDate: time.Now(),
    Addresses: map[string]*thirdparty.Address{
        "ShipFrom": {Country: "US", Region: "CA", ...},
        "ShipTo":   {Country: "US", Region: "NY", ...},
    },
    Lines: []*thirdparty.LineItem{
        {Amount: 100.0, TaxCode: "P0000000"},
    },
}
response, err := provider.CalculateTax(ctx, request)
```

## 4. 税务报表生成

位置：`internal/tax/infrastructure/reporting/`

### 4.1 支持的报表类型

#### 增值税申报表 (VAT Return)
- 英国/欧盟VAT申报格式(Box 1-9)
- 销项税额汇总
- 进项税额汇总
- 净应缴或应退税额计算

#### 消费税申报表 (Excise Return)
- 按商品类别汇总
- 从价税和从量税分别统计
- 消费税明细

#### 关税申报表 (Customs Duty Return)
- 进口报关单统计
- 关税、进口增值税、进口消费税汇总
- HS编码分类统计

#### 综合税务报表
- 多税种合并报表
- 按期间汇总
- 按税种分类统计

### 4.2 支持的报表格式
- **JSON**: 结构化数据，便于系统集成
- **CSV**: Excel兼容，便于人工查看
- **PDF**: 正式申报文档（需扩展实现）
- **Excel**: 电子表格格式（需扩展实现）
- **XML**: 标准数据交换格式

### 4.3 使用方式
```go
// 生成增值税申报表
reportData, err := taxService.GenerateVATReturn(
    ctx,
    startDate,      // 2024-01-01
    endDate,        // 2024-03-31
    "GB",           // 国家代码
    true,           // 包含明细
)

// 导出CSV格式
request := &reporting.ReportRequest{
    ReportType:     reporting.ReportTypeVATReturn,
    Format:         reporting.FormatCSV,
    StartDate:      startDate,
    EndDate:        endDate,
    CountryCode:    "GB",
    IncludeDetails: true,
}
var buf bytes.Buffer
err = taxService.ExportReport(ctx, request, &buf)
```

## 5. 应用服务层集成

位置：`internal/tax/application/service.go`

### 5.1 核心服务方法

#### 税务计算
- `CalculateOrderTax`: 基础订单税费计算
- `CalculateVAT`: 增值税专项计算
- `CalculateExcise`: 消费税专项计算
- `CalculateDuty`: 关税专项计算
- `CalculateComprehensiveTax`: 综合税费计算
- `CalculateCrossBorderTax`: 跨境税务计算
- `CalculateTaxWithThirdParty`: 使用第三方服务计算

#### 报表生成
- `GenerateVATReturn`: 生成增值税申报表
- `GenerateExciseReturn`: 生成消费税申报表
- `GenerateDutyReturn`: 生成关税申报表
- `GenerateConsolidatedReport`: 生成综合税务报表

#### 地址和税率
- `ValidateAddress`: 验证地址（第三方服务）
- `GetTaxRates`: 获取税率（第三方服务）

#### 配置管理
- `SaveCrossBorderConfig`: 保存跨境税务配置
- `GetCrossBorderConfig`: 获取跨境税务配置
- `GetPreferentialAgreements`: 获取优惠贸易协定
- `CreateTaxRule`: 创建税务规则

#### 免税管理
- `ApplyTaxExemption`: 申请税务减免
- `GetTaxExemption`: 获取用户免税状态

## 6. 文件结构

```
ecommerce/internal/tax/
├── domain/
│   ├── tax.go                  # 基础税务领域模型
│   ├── tax_calculator.go       # 通用税务计算器
│   ├── calculators.go          # 专项税务计算引擎(VAT/Excise/Duty)
│   └── crossborder.go          # 跨境税务规则和配置
├── application/
│   └── service.go              # 应用服务（扩展后）
├── infrastructure/
│   ├── persistence/
│   │   └── mysql/
│   │       ├── models.go       # 数据库模型
│   │       └── repository.go   # 仓储实现
│   ├── thirdparty/
│   │   ├── provider.go         # 第三方服务接口和工厂
│   │   ├── avalara.go          # Avalara集成
│   │   └── vertex.go           # Vertex集成
│   └── reporting/
│       └── report_generator.go # 税务报表生成器
└── interfaces/
    └── grpc/
        └── server.go           # gRPC接口
```

## 7. 使用示例

### 7.1 计算跨境电商订单税费
```go
ctx := context.Background()

// 配置跨境税务计算
input := &domain.CrossBorderTaxInput{
    OriginCountry:      "CN",
    DestinationCountry: "US",
    TradeType:          domain.TradeTypeB2C,
    TransactionMode:    domain.TransactionModeDDP,
    HSCode:             "8517.12.00",
    Category:           "ELECTRONICS",
    ProductValue:       50000,     // $500
    ShippingCost:       1500,      // $15
    InsuranceCost:      500,       // $5
    Quantity:           1,
    Weight:             0.5,
    HasOriginCert:      false,
}

result, err := taxService.CalculateCrossBorderTax(ctx, input)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("关税: $%.2f\n", float64(result.CustomsDuty)/100)
fmt.Printf("进口增值税: $%.2f\n", float64(result.ImportVAT)/100)
fmt.Printf("总税费: $%.2f\n", float64(result.TotalImportTax)/100)
fmt.Printf("所需单证: %v\n", result.RequiredDocuments)
```

### 7.2 使用Avalara计算美国销售税
```go
// 配置Avalara提供商
avalaraConfig := &thirdparty.ProviderConfig{
    APIBaseURL:  "https://rest.avatax.com",
    APIKey:      base64.StdEncoding.EncodeToString([]byte("account:license")),
    CompanyCode: "DEFAULT",
}

factory := thirdparty.NewProviderFactory()
factory.RegisterConfig("avalara", avalaraConfig)
provider, _ := factory.CreateProvider("avalara")

// 配置税务服务
taxConfig := &application.TaxServiceConfig{
    UseThirdPartyProvider: true,
    ThirdPartyProvider:    provider,
}
taxService := application.NewTaxService(repo, crossBorderRepo, taxConfig, logger)

// 计算税费
request := &thirdparty.TaxCalculationRequest{
    TransactionDate: time.Now(),
    CustomerCode:    "CUST001",
    DocumentType:    "SalesInvoice",
    Addresses: map[string]*thirdparty.Address{
        "ShipFrom": {
            Line1:   "123 Main St",
            City:    "Seattle",
            Region:  "WA",
            Country: "US",
            PostalCode: "98101",
        },
        "ShipTo": {
            Line1:   "456 Oak Ave",
            City:    "New York",
            Region:  "NY",
            Country: "US",
            PostalCode: "10001",
        },
    },
    Lines: []*thirdparty.LineItem{
        {
            LineNo:   "1",
            ItemCode: "SKU001",
            TaxCode:  "P0000000",
            Quantity: 2,
            Amount:   200.00,
        },
    },
}

response, err := taxService.CalculateTaxWithThirdParty(ctx, request)
```

### 7.3 生成季度增值税申报表
```go
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
endDate := time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)

report, err := taxService.GenerateVATReturn(ctx, startDate, endDate, "GB", true)
if err != nil {
    log.Fatal(err)
}

// 保存CSV文件
request := &reporting.ReportRequest{
    ReportType:     reporting.ReportTypeVATReturn,
    Format:         reporting.FormatCSV,
    StartDate:      startDate,
    EndDate:        endDate,
    CountryCode:    "GB",
    IncludeDetails: true,
}

file, _ := os.Create("vat_return_q1_2024.csv")
defer file.Close()
err = taxService.ExportReport(ctx, request, file)
```

## 8. 扩展指南

### 8.1 添加新的第三方税务服务
1. 在 `thirdparty/` 目录下创建新的提供商文件（如 `taxjar.go`）
2. 实现 `ThirdPartyTaxProvider` 接口
3. 在 `ProviderFactory.CreateProvider()` 中添加新的 case

### 8.2 添加新的报表类型
1. 在 `reporting/` 目录下实现新的 Report 接口
2. 在 `ReportGenerator.Generate()` 中添加新的 case
3. 更新 `GetSupportedReportTypes()`

### 8.3 添加新的税务计算规则
1. 在 `domain/calculators.go` 中添加新的计算器
2. 在 `ComprehensiveTaxCalculator` 中集成
3. 在 `TaxService` 中添加对应的公共方法

## 9. 注意事项

1. **货币单位**: 系统内部使用"分"作为货币单位，对外展示时转换为元
2. **税率精度**: 税率存储为小数形式（如 0.20 表示 20%）
3. **时区处理**: 所有时间使用 UTC，展示时根据需要进行时区转换
4. **第三方服务**: 生产环境需要配置正确的API密钥和基础URL
5. **报表格式**: PDF和Excel格式需要额外实现（当前只有JSON和CSV）
