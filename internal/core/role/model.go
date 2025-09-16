package role

type MstRole struct {
	RoleId      uint            `json:"role_id" gorm:"primaryKey;autoIncrement"`
	Name        string          `json:"name"`
	Permissions []MstPermission `json:"permissions" gorm:"many2many:ref_role_permissions"`
}

type MstPermission struct {
	PermissionId uint   `json:"permission_id" gorm:"primaryKey;autoIncrement"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
}

type RoleFilter struct {
	Name string
}
