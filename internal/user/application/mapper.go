package application

import "github.com/wyfcoding/ecommerce/internal/user/domain"

func toUserDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toAddressDTO(addr *domain.Address) *AddressDTO {
	if addr == nil {
		return nil
	}
	return &AddressDTO{
		ID:              addr.ID,
		UserID:          addr.UserID,
		RecipientName:   addr.RecipientName,
		PhoneNumber:     addr.PhoneNumber,
		Province:        addr.Province,
		City:            addr.City,
		District:        addr.District,
		DetailedAddress: addr.DetailedAddress,
		PostalCode:      addr.PostalCode,
		IsDefault:       addr.IsDefault,
		CreatedAt:       addr.CreatedAt,
		UpdatedAt:       addr.UpdatedAt,
	}
}

func toAddressDTOs(addrs []*domain.Address) []*AddressDTO {
	if len(addrs) == 0 {
		return nil
	}
	dtos := make([]*AddressDTO, len(addrs))
	for i, addr := range addrs {
		dtos[i] = toAddressDTO(addr)
	}
	return dtos
}
