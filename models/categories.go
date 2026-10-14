package models

type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"uniqueIndex;not null;size:32"`
	Name string `gorm:"not null;size:256"`
}

func (c *Category) TableName() string {
	return "categories"
}
