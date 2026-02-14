package interfaces

import (
	"context"
	"encoding/json"
	"fmt"

	merchantv1 "github.com/wyfcoding/ecommerce/go-api/merchant/v1"
	"github.com/wyfcoding/ecommerce/internal/merchant/application"
	"github.com/wyfcoding/ecommerce/internal/merchant/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GRPCHandler) Apply(ctx context.Context, req *merchantv1.ApplyRequest) (*merchantv1.ApplyResponse, error) {
	cmd := application.ApplyCommand{
		UserID:       req.GetUserId(),
		Name:         req.GetName(),
		LegalName:    req.GetLegalName(),
		LegalIDCard:  req.GetLegalIdCard(),
		ContactName:  req.GetContactName(),
		ContactPhone: req.GetContactPhone(),
		ContactEmail: req.GetContactEmail(),
		Type:         domain.MerchantType(req.GetType()),
		LogoURL:      req.GetLogoUrl(),
		Description:  req.GetDescription(),
	}
	if req.GetBusinessLicense() != nil {
		cmd.License = &application.LicenseDTO{
			LicenseNo:         req.GetBusinessLicense().GetLicenseNo(),
			CompanyName:       req.GetBusinessLicense().GetCompanyName(),
			CompanyType:       req.GetBusinessLicense().GetCompanyType(),
			RegisteredCapital: req.GetBusinessLicense().GetRegisteredCapital(),
			EstablishmentDate: req.GetBusinessLicense().GetEstablishmentDate(),
			ValidUntil:        req.GetBusinessLicense().GetValidUntil(),
			BusinessScope:     req.GetBusinessLicense().GetBusinessScope(),
			RegisteredAddress: req.GetBusinessLicense().GetRegisteredAddress(),
			LicenseImageURL:   req.GetBusinessLicense().GetLicenseImageUrl(),
		}
	}
	if req.GetBankAccount() != nil {
		cmd.BankAccount = &application.BankAccountDTO{
			AccountName: req.GetBankAccount().GetAccountName(),
			AccountNo:   req.GetBankAccount().GetAccountNo(),
			BankName:    req.GetBankAccount().GetBankName(),
			BankBranch:  req.GetBankAccount().GetBankBranch(),
			BankCode:    req.GetBankAccount().GetBankCode(),
		}
	}

	result, err := h.commandService.Apply(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &merchantv1.ApplyResponse{
		MerchantId: uint64(result.MerchantID),
		MerchantNo: result.MerchantNo,
	}, nil
}

func (h *GRPCHandler) Approve(ctx context.Context, req *merchantv1.ApproveRequest) (*emptypb.Empty, error) {
	err := h.commandService.Approve(ctx, application.ApproveCommand{
		MerchantID:     uint(req.GetMerchantId()),
		CommissionRate: req.GetCommissionRate(),
		Operator:       req.GetOperator(),
		Remark:         req.GetRemark(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *GRPCHandler) Reject(ctx context.Context, req *merchantv1.RejectRequest) (*emptypb.Empty, error) {
	err := h.commandService.Reject(ctx, application.RejectCommand{
		MerchantID: uint(req.GetMerchantId()),
		Reason:     req.GetReason(),
		Operator:   req.GetOperator(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *GRPCHandler) Update(ctx context.Context, req *merchantv1.UpdateMerchantRequest) (*merchantv1.Merchant, error) {
	cmd := application.UpdateMerchantCommand{
		MerchantID:   uint(req.GetMerchantId()),
		Name:         req.Name,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		LogoURL:      req.LogoUrl,
		Description:  req.Description,
	}
	if req.GetBankAccount() != nil {
		cmd.BankAccount = &application.BankAccountDTO{
			AccountName: req.GetBankAccount().GetAccountName(),
			AccountNo:   req.GetBankAccount().GetAccountNo(),
			BankName:    req.GetBankAccount().GetBankName(),
			BankBranch:  req.GetBankAccount().GetBankBranch(),
			BankCode:    req.GetBankAccount().GetBankCode(),
		}
	}
	if err := h.commandService.UpdateMerchant(ctx, cmd); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	merchant, err := h.queryService.GetMerchant(ctx, uint(req.GetMerchantId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoMerchant(merchant), nil
}

func (h *GRPCHandler) Get(ctx context.Context, req *merchantv1.GetMerchantRequest) (*merchantv1.Merchant, error) {
	merchant, err := h.queryService.GetMerchant(ctx, uint(req.GetMerchantId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoMerchant(merchant), nil
}

func (h *GRPCHandler) GetByUserID(ctx context.Context, req *merchantv1.GetByUserIDRequest) (*merchantv1.Merchant, error) {
	merchant, err := h.queryService.GetMerchantByUserID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoMerchant(merchant), nil
}

func (h *GRPCHandler) List(ctx context.Context, req *merchantv1.ListMerchantsRequest) (*merchantv1.ListMerchantsResponse, error) {
	page := max(int(req.GetPage()), 1)
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	filter := &domain.MerchantFilter{
		Page:     page,
		PageSize: pageSize,
		Keyword:  req.GetKeyword(),
	}
	if req.Status != nil && req.GetStatus() != merchantv1.MerchantStatus_MERCHANT_STATUS_UNSPECIFIED {
		statusVal := domain.MerchantStatus(req.GetStatus())
		filter.Status = &statusVal
	}
	if req.Type != nil && req.GetType() != merchantv1.MerchantType_MERCHANT_TYPE_UNSPECIFIED {
		typeVal := domain.MerchantType(req.GetType())
		filter.Type = &typeVal
	}
	if req.Level != nil && req.GetLevel() != merchantv1.MerchantLevel_MERCHANT_LEVEL_UNSPECIFIED {
		levelVal := domain.MerchantLevel(req.GetLevel())
		filter.Level = &levelVal
	}

	result, err := h.queryService.ListMerchants(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	merchants := make([]*merchantv1.Merchant, 0, len(result.Items))
	for _, item := range result.Items {
		merchants = append(merchants, toProtoMerchant(item))
	}

	return &merchantv1.ListMerchantsResponse{
		Merchants: merchants,
		Total:     int32(result.Total),
	}, nil
}

func (h *GRPCHandler) Disable(ctx context.Context, req *merchantv1.DisableMerchantRequest) (*emptypb.Empty, error) {
	err := h.commandService.Disable(ctx, application.DisableCommand{
		MerchantID: uint(req.GetMerchantId()),
		Reason:     req.GetReason(),
		Operator:   req.GetOperator(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *GRPCHandler) Enable(ctx context.Context, req *merchantv1.EnableMerchantRequest) (*emptypb.Empty, error) {
	err := h.commandService.Enable(ctx, application.EnableCommand{
		MerchantID: uint(req.GetMerchantId()),
		Operator:   req.GetOperator(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *GRPCHandler) UpdateSettings(ctx context.Context, req *merchantv1.UpdateSettingsRequest) (*merchantv1.MerchantSettings, error) {
	if req.GetSettings() == nil {
		return nil, status.Error(codes.InvalidArgument, "settings is required")
	}
	settings := req.GetSettings()
	cmd := application.UpdateSettingsCommand{
		MerchantID: uint(req.GetMerchantId()),
		Settings: &application.SettingsDTO{
			AutoConfirmOrder:       new(settings.GetAutoConfirmOrder()),
			AutoConfirmDays:        new(settings.GetAutoConfirmDays()),
			EnableCOD:              new(settings.GetEnableCod()),
			EnableInvoice:          new(settings.GetEnableInvoice()),
			SupportedPayments:      settings.GetSupportedPayments(),
			FreeShippingThreshold:  new(settings.GetShippingSettings().GetFreeShippingThreshold()),
			DefaultShippingFee:     new(settings.GetShippingSettings().GetDefaultShippingFee()),
			EnableSameDayDelivery:  new(settings.GetShippingSettings().GetEnableSameDayDelivery()),
			OrderNotification:      new(settings.GetNotificationSettings().GetOrderNotification()),
			StockAlert:             new(settings.GetNotificationSettings().GetStockAlert()),
			ReviewNotification:     new(settings.GetNotificationSettings().GetReviewNotification()),
			SettlementNotification: new(settings.GetNotificationSettings().GetSettlementNotification()),
			NotificationPhone:      new(settings.GetNotificationSettings().GetNotificationPhone()),
			NotificationEmail:      new(settings.GetNotificationSettings().GetNotificationEmail()),
		},
	}

	updated, err := h.commandService.UpdateSettings(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoSettings(updated), nil
}

func (h *GRPCHandler) GetSettings(ctx context.Context, req *merchantv1.GetSettingsRequest) (*merchantv1.MerchantSettings, error) {
	settings, err := h.queryService.GetSettings(ctx, uint(req.GetMerchantId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoSettings(settings), nil
}

func (h *GRPCHandler) CreateStore(ctx context.Context, req *merchantv1.CreateStoreRequest) (*merchantv1.Store, error) {
	store, err := h.commandService.CreateStore(ctx, application.CreateStoreCommand{
		MerchantID:    uint(req.GetMerchantId()),
		Name:          req.GetName(),
		LogoURL:       req.GetLogoUrl(),
		BannerURL:     req.GetBannerUrl(),
		Description:   req.GetDescription(),
		Categories:    req.GetCategories(),
		BusinessHours: req.GetBusinessHours(),
		Address:       req.GetAddress(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoStore(&application.StoreDTO{
		ID:            store.ID,
		CreatedAt:     store.CreatedAt,
		UpdatedAt:     store.UpdatedAt,
		MerchantID:    store.MerchantID,
		StoreNo:       store.StoreNo,
		Name:          store.Name,
		LogoURL:       store.LogoURL,
		BannerURL:     store.BannerURL,
		Description:   store.Description,
		Announcement:  store.Announcement,
		Categories:    store.Categories,
		Rating:        store.Rating,
		ProductCount:  store.ProductCount,
		MonthlySales:  store.MonthlySales,
		IsOpen:        store.IsOpen,
		BusinessHours: store.BusinessHours,
		Address:       store.Address,
	}), nil
}

func (h *GRPCHandler) UpdateStore(ctx context.Context, req *merchantv1.UpdateStoreRequest) (*merchantv1.Store, error) {
	store, err := h.commandService.UpdateStore(ctx, application.UpdateStoreCommand{
		StoreID:       uint(req.GetStoreId()),
		Name:          req.Name,
		LogoURL:       req.LogoUrl,
		BannerURL:     req.BannerUrl,
		Description:   req.Description,
		Announcement:  req.Announcement,
		Categories:    req.GetCategories(),
		IsOpen:        req.IsOpen,
		BusinessHours: req.BusinessHours,
		Address:       req.Address,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoStore(&application.StoreDTO{
		ID:            store.ID,
		CreatedAt:     store.CreatedAt,
		UpdatedAt:     store.UpdatedAt,
		MerchantID:    store.MerchantID,
		StoreNo:       store.StoreNo,
		Name:          store.Name,
		LogoURL:       store.LogoURL,
		BannerURL:     store.BannerURL,
		Description:   store.Description,
		Announcement:  store.Announcement,
		Categories:    store.Categories,
		Rating:        store.Rating,
		ProductCount:  store.ProductCount,
		MonthlySales:  store.MonthlySales,
		IsOpen:        store.IsOpen,
		BusinessHours: store.BusinessHours,
		Address:       store.Address,
	}), nil
}

func (h *GRPCHandler) GetStore(ctx context.Context, req *merchantv1.GetStoreRequest) (*merchantv1.Store, error) {
	store, err := h.queryService.GetStore(ctx, uint(req.GetStoreId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoStore(store), nil
}

func (h *GRPCHandler) ListStores(ctx context.Context, req *merchantv1.ListStoresRequest) (*merchantv1.ListStoresResponse, error) {
	page := max(int(req.GetPage()), 1)
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 20
	}

	result, err := h.queryService.ListStores(ctx, uint(req.GetMerchantId()), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	stores := make([]*merchantv1.Store, 0, len(result.Items))
	for _, item := range result.Items {
		stores = append(stores, toProtoStore(item))
	}
	return &merchantv1.ListStoresResponse{
		Stores: stores,
		Total:  int32(result.Total),
	}, nil
}

func toProtoMerchant(dto *application.MerchantDTO) *merchantv1.Merchant {
	if dto == nil {
		return nil
	}
	resp := &merchantv1.Merchant{
		Id:             uint64(dto.ID),
		UserId:         dto.UserID,
		MerchantNo:     dto.MerchantNo,
		Name:           dto.Name,
		LegalName:      dto.LegalName,
		ContactName:    dto.ContactName,
		ContactPhone:   dto.ContactPhone,
		ContactEmail:   dto.ContactEmail,
		Type:           merchantv1.MerchantType(dto.Type),
		Status:         merchantv1.MerchantStatus(dto.Status),
		Level:          merchantv1.MerchantLevel(dto.Level),
		LogoUrl:        dto.LogoURL,
		Description:    dto.Description,
		CommissionRate: dto.CommissionRate,
		CreditScore:    dto.CreditScore,
		TotalSales:     dto.TotalSales,
		TotalOrders:    dto.TotalOrders,
		Rating:         dto.Rating,
		RejectReason:   dto.RejectReason,
		CreatedAt:      timestamppb.New(dto.CreatedAt),
		UpdatedAt:      timestamppb.New(dto.UpdatedAt),
	}
	if dto.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*dto.ApprovedAt)
	}
	if dto.BusinessLicense != nil {
		resp.BusinessLicense = &merchantv1.BusinessLicense{
			LicenseNo:         dto.BusinessLicense.LicenseNo,
			CompanyName:       dto.BusinessLicense.CompanyName,
			CompanyType:       dto.BusinessLicense.CompanyType,
			RegisteredCapital: dto.BusinessLicense.RegisteredCapital,
			EstablishmentDate: dto.BusinessLicense.EstablishmentDate,
			ValidUntil:        dto.BusinessLicense.ValidUntil,
			BusinessScope:     dto.BusinessLicense.BusinessScope,
			RegisteredAddress: dto.BusinessLicense.RegisteredAddress,
			LicenseImageUrl:   dto.BusinessLicense.LicenseImageURL,
		}
	}
	if dto.BankAccount != nil {
		resp.BankAccount = &merchantv1.BankAccount{
			AccountName: dto.BankAccount.AccountName,
			AccountNo:   dto.BankAccount.AccountNo,
			BankName:    dto.BankAccount.BankName,
			BankBranch:  dto.BankAccount.BankBranch,
			BankCode:    dto.BankAccount.BankCode,
		}
	}
	return resp
}

func toProtoSettings(settings *domain.MerchantSettings) *merchantv1.MerchantSettings {
	if settings == nil {
		return &merchantv1.MerchantSettings{}
	}

	var supportedPayments []string
	if settings.SupportedPayments != "" {
		_ = json.Unmarshal([]byte(settings.SupportedPayments), &supportedPayments)
	}

	return &merchantv1.MerchantSettings{
		MerchantId:        uint64(settings.MerchantID),
		AutoConfirmOrder:  settings.AutoConfirmOrder,
		AutoConfirmDays:   settings.AutoConfirmDays,
		EnableCod:         settings.EnableCOD,
		EnableInvoice:     settings.EnableInvoice,
		SupportedPayments: supportedPayments,
		ShippingSettings: &merchantv1.ShippingSettings{
			SupportedCarrierIds:   []uint64{},
			FreeShippingThreshold: settings.FreeShippingThreshold,
			DefaultShippingFee:    settings.DefaultShippingFee,
			EnableSameDayDelivery: settings.EnableSameDayDelivery,
		},
		NotificationSettings: &merchantv1.NotificationSettings{
			OrderNotification:      settings.OrderNotification,
			StockAlert:             settings.StockAlert,
			ReviewNotification:     settings.ReviewNotification,
			SettlementNotification: settings.SettlementNotification,
			NotificationPhone:      settings.NotificationPhone,
			NotificationEmail:      settings.NotificationEmail,
		},
	}
}

func toProtoStore(dto *application.StoreDTO) *merchantv1.Store {
	if dto == nil {
		return nil
	}

	categories, err := decodeCategories(dto.Categories)
	if err != nil {
		categories = []string{}
	}

	return &merchantv1.Store{
		Id:            uint64(dto.ID),
		MerchantId:    uint64(dto.MerchantID),
		StoreNo:       dto.StoreNo,
		Name:          dto.Name,
		LogoUrl:       dto.LogoURL,
		BannerUrl:     dto.BannerURL,
		Description:   dto.Description,
		Announcement:  dto.Announcement,
		Categories:    categories,
		Rating:        dto.Rating,
		ProductCount:  dto.ProductCount,
		MonthlySales:  dto.MonthlySales,
		IsOpen:        dto.IsOpen,
		BusinessHours: dto.BusinessHours,
		Address:       dto.Address,
		CreatedAt:     timestamppb.New(dto.CreatedAt),
		UpdatedAt:     timestamppb.New(dto.UpdatedAt),
	}
}

func decodeCategories(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}
	var categories []string
	if err := json.Unmarshal([]byte(raw), &categories); err != nil {
		return nil, fmt.Errorf("invalid categories json: %w", err)
	}
	return categories, nil
}

//go:fix inline
func boolPtr(v bool) *bool { return new(v) }

//go:fix inline
func int32Ptr(v int32) *int32 { return new(v) }

//go:fix inline
func int64Ptr(v int64) *int64 { return new(v) }

//go:fix inline
func stringPtr(v string) *string { return new(v) }
