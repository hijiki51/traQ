package model

import (
	"github.com/satori/go.uuid"
	"github.com/traPtitech/traQ/utils/validator"
	"time"
)

// Message データベースに格納するmessageの構造体
type Message struct {
	ID        uuid.UUID  `gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID  `gorm:"type:char(36)"`
	ChannelID uuid.UUID  `gorm:"type:char(36);index"`
	Text      string     `gorm:"type:text"                 validate:"required"`
	CreatedAt time.Time  `gorm:"precision:6;index"`
	UpdatedAt time.Time  `gorm:"precision:6"`
	DeletedAt *time.Time `gorm:"precision:6;index"`
}

// TableName DBの名前を指定するメソッド
func (m *Message) TableName() string {
	return "messages"
}

// Validate 構造体を検証します
func (m *Message) Validate() error {
	return validator.ValidateStruct(m)
}

// Unread 未読レコード
type Unread struct {
	UserID    uuid.UUID `gorm:"type:char(36);primary_key"`
	MessageID uuid.UUID `gorm:"type:char(36);primary_key"`
	CreatedAt time.Time `gorm:"precision:6"`
}

// TableName テーブル名
func (unread *Unread) TableName() string {
	return "unreads"
}

// ArchivedMessage 編集前のアーカイブ化されたメッセージの構造体
type ArchivedMessage struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	MessageID uuid.UUID `gorm:"type:char(36);index"`
	UserID    uuid.UUID `gorm:"type:char(36)"`
	Text      string    `gorm:"type:text"`
	DateTime  time.Time `gorm:"precision:6"`
}

// TableName ArchivedMessage構造体のテーブル名
func (am *ArchivedMessage) TableName() string {
	return "archived_messages"
}
