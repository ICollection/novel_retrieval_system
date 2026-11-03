package service

// AccountService 系统用户服务。
type AccountService struct {
}

// Append 添加系统用户。
func (service *AccountService) Append(name *string, login *string, password *string, permission *string) (*string, error) {
	return nil, nil
}

// Modify 修改系统用户。
func (service *AccountService) Modify(id *string, name *string, permission *string) (*string, error) {
	return id, nil
}

// Change 改密系统用户。
func (service *AccountService) Change(id *string, password *string) (*string, error) {
	return id, nil
}

// Delete 删除系统用户。
func (service *AccountService) Delete(id *string) (*string, error) {
	return id, nil
}

// Verify 验证系统用户。
func (service *AccountService) Verify(login *string, password *string) (*string, error) {
	return nil, nil
}

// Filter 查询系统用户，
func (service *AccountService) Filter() ([]Account, error) {
	return nil, nil
}

// Account 系统用户。
type Account struct {
	Id         *string `db:"id" json:"id"`
	Name       *string `db:"name" json:"name"`
	Login      *string `db:"login" json:"login"`
	Password   *string `db:"password" json:"password"`
	Permission *string `db:"permission" json:"permission"`
}
