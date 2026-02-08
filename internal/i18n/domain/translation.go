package domain

import (
	"gorm.io/gorm"
)

// Language 语言定义
type Language struct {
	gorm.Model
	Code      string `gorm:"column:code;type:varchar(10);unique_index;not null"` // zh-CN, en-US
	Name      string `gorm:"column:name;type:varchar(50);not null"`              // Simplified Chinese
	Direction string `gorm:"column:direction;type:varchar(3);default:'LTR'"`     // LTR, RTL
	Enabled   bool   `gorm:"column:enabled;not null;default:true"`
}

// Translation 翻译条目
type Translation struct {
	gorm.Model
	LangCode  string `gorm:"column:lang_code;type:varchar(10);unique_index:idx_lang_key;not null"`
	Key       string `gorm:"column:key;type:varchar(255);unique_index:idx_lang_key;not null"`
	Value     string `gorm:"column:value;type:text;not null"`
	Context   string `gorm:"column:context;type:varchar(50)"`   // 上下文区分，如 "menu", "button"
	Namespace string `gorm:"column:namespace;type:varchar(50)"` // 命名空间/模块，如 "procurement", "common"
}

func (Language) TableName() string    { return "languages" }
func (Translation) TableName() string { return "translations" }

func NewLanguage(code, name, dir string) *Language {
	return &Language{
		Code:      code,
		Name:      name,
		Direction: dir,
		Enabled:   true,
	}
}

func NewTranslation(lang, key, value, context, ns string) *Translation {
	return &Translation{
		LangCode:  lang,
		Key:       key,
		Value:     value,
		Context:   context,
		Namespace: ns,
	}
}
