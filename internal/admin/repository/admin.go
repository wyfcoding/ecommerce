package repository

import (
	"context"
	"ecommerce/internal/admin/biz"
	"ecommerce/internal/admin/model"

	productV1 "ecommerce/api/product/v1"

	"gorm.io/gorm"
	"google.golang.org/grpc"
)

type adminRepo struct {
	data *Data
}

// NewAdminRepo 创建一个新的 AdminRepo。
func NewAdminRepo(data *Data) biz.AdminRepo {
	return &adminRepo{data: data}
}

// toBizAdminUser 将 data.AdminUser 转换为 biz.AdminUser。
func (r *adminRepo) toBizAdminUser(po *model.AdminUser) *biz.AdminUser {
	if po == nil {
		return nil
	}
	return &biz.AdminUser{
		ID:       uint32(po.ID),
		Username: po.Username,
		Password: po.Password,
		Name:     po.Name,
		Status:   po.Status,
	}
}

// CreateAdminUser 创建一个新的管理员用户。
func (r *adminRepo) CreateAdminUser(ctx context.Context, user *biz.AdminUser) (*biz.AdminUser, error) {
	po := &model.AdminUser{
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	user.ID = uint32(po.ID) // 将生成的 ID 赋值回 biz.AdminUser。
	return r.toBizAdminUser(po), nil
}

// GetAdminUserByUsername 根据用户名获取管理员用户。
func (r *adminRepo) GetAdminUserByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var po model.AdminUser
	if err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 用户未找到
		}
		return nil, err
	}
	return r.toBizAdminUser(&po), nil
}

type authRepo struct {
	*Data
}

// NewAuthRepo 是 authRepo 的构造函数。
func NewAuthRepo(data *Data) biz.AuthRepo {
	return &authRepo{Data: data}
}

// GetAdminUserByUsername 根据用户名从数据仓库中获取管理员用户信息。
func (r *authRepo) GetAdminUserByUsername(ctx context.Context, username string) (*biz.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.AdminUser{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}, nil
}

// GetAdminUserByID 根据ID从数据仓库中获取管理员用户信息。
func (r *authRepo) GetAdminUserByID(ctx context.Context, id uint32) (*biz.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.AdminUser{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Name:     user.Name,
		Status:   user.Status,
	}, nil
}

// Data 结构体持有数据库连接。
type Data struct {
	db *gorm.DB
}

// NewData 是 Data 结构体的构造函数。
func NewData(db *gorm.DB) *Data {
	return &Data{db: db}
}

// transaction 实现了 biz.Transaction 接口。
type transaction struct {
	db *gorm.DB
}

// NewTransaction 是 transaction 的构造函数。
func NewTransaction(data *Data) biz.Transaction {
	return &transaction{db: data.db}
}

// InTx 在一个数据库事务中执行函数。
func (t *transaction) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}

// productClient 实现了 biz.ProductClient 接口。
type productClient struct {
	client productV1.ProductServiceClient
}

// NewProductClient 是 productClient 的构造函数。
func NewProductClient(conn *grpc.ClientConn) biz.ProductClient {
	return &productClient{client: productV1.NewProductServiceClient(conn)}
}

// protoToBizSpu 将 productV1.SpuInfo 转换为 biz.Spu。
func (pc *productClient) protoToBizSpu(spu *productV1.SpuInfo) *biz.Spu {
	if spu == nil {
		return nil
	}
	return &biz.Spu{
		ID:            spu.SpuId,
		CategoryID:    spu.CategoryId,
		BrandID:       spu.BrandId,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHTML:    spu.DetailHTML,
		Status:        spu.Status,
	}
}

// bizSpuToProto 将 biz.Spu 转换为 productV1.SpuInfo。
func (pc *productClient) bizSpuToProto(spu *biz.Spu) *productV1.SpuInfo {
	if spu == nil {
		return nil
	}
	res := &productV1.SpuInfo{
		SpuId:         spu.ID,
		CategoryId:    spu.CategoryID,
		BrandId:       spu.BrandID,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHtml:    spu.DetailHTML,
		Status:        spu.Status,
	}
	return res
}

// protoToBizSku 将 productV1.SkuInfo 转换为 biz.Sku。
func (pc *productClient) protoToBizSku(sku *productV1.SkuInfo) *biz.Sku {
	if sku == nil {
		return nil
	}
	return &biz.Sku{
		ID:            sku.SkuId,
		SpuID:         sku.SpuId,
		Title:         sku.Title,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		Stock:         sku.Stock,
		Image:         sku.Image,
		Specs:         sku.Specs,
		Status:        sku.Status,
	}
}

// bizSkuToProto 将 biz.Sku 转换为 productV1.SkuInfo。
func (pc *productClient) bizSkuToProto(sku *biz.Sku) *productV1.SkuInfo {
	if sku == nil {
		return nil
	}
	res := &productV1.SkuInfo{
		SkuId:         sku.ID,
		SpuId:         sku.SpuID,
		Title:         sku.Title,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		Stock:         sku.Stock,
		Image:         sku.Image,
		Specs:         sku.Specs,
		Status:        sku.Status,
	}
	return res
}

// protoToBizSkuList 将 productV1.SkuInfo 列表转换为 biz.Sku 列表。
func (pc *productClient) protoToBizSkuList(skus []*productV1.SkuInfo) []*biz.Sku {
	if skus == nil {
		return nil
	}
	bizSkus := make([]*biz.Sku, 0, len(skus))
	for _, skuProto := range skus {
		bizSkus = append(bizSkus, pc.protoToBizSku(skuProto))
	}
	return bizSkus
}

// CreateProduct 调用商品服务创建商品。
func (pc *productClient) CreateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	protoSpu := pc.bizSpuToProto(spu)
	protoSkus := make([]*productV1.SkuInfo, 0, len(skus))
	for _, sku := range skus {
		protoSkus = append(protoSkus, pc.bizSkuToProto(sku))
	}

	res, err := pc.client.CreateProduct(ctx, &productV1.CreateProductRequest{
		Spu:  protoSpu,
		Skus: protoSkus,
	})
	if err != nil {
		return nil, nil, err
	}

	return pc.protoToBizSpu(res.Spu), pc.protoToBizSkuList(res.Skus), nil
}

// UpdateProduct 调用商品服务更新商品。
func (pc *productClient) UpdateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	protoSpu := pc.bizSpuToProto(spu)
	protoSkus := make([]*productV1.SkuInfo, 0, len(skus))
	for _, sku := range skus {
		protoSkus = append(protoSkus, pc.bizSkuToProto(sku))
	}

	res, err := pc.client.UpdateProduct(ctx, &productV1.UpdateProductRequest{
		Spu:  protoSpu,
		Skus: protoSkus,
	})
	if err != nil {
		return nil, nil, err
	}

	// 假设 UpdateProduct 返回更新后的完整 SPU 和 SKU 信息
	updatedSpu := pc.protoToBizSpu(res.Spu)
	updatedSkus := make([]*biz.Sku, 0, len(res.Skus))
	for _, skuProto := range res.Skus {
		updatedSkus = append(updatedSkus, pc.protoToBizSku(skuProto))
	}

	return updatedSpu, updatedSkus, nil
}

// GetSpuDetail 调用商品服务获取商品详情。
func (pc *productClient) GetSpuDetail(ctx context.Context, spuID uint64) (*biz.Spu, []*biz.Sku, error) {
	res, err := pc.client.GetSpuDetail(ctx, &productV1.GetSpuDetailRequest{SpuId: spuID})
	if err != nil {
		return nil, nil, err
	}

	bizSpu := pc.protoToBizSpu(res.Spu)
	bizSkus := make([]*biz.Sku, 0, len(res.Skus))
	for _, skuProto := range res.Skus {
		bizSkus = append(bizSkus, pc.protoToBizSku(skuProto))
	}

	return bizSpu, bizSkus, nil
}