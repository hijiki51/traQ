package repository

import (
	"github.com/gofrs/uuid"

	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/utils/optional"
)

// ClipFolderMessageQuery クリップフォルダー内のメッセージ取得用クエリ
type ClipFolderMessageQuery struct {
	Limit  int
	Offset int
	Asc    bool
}

// ClipRepository クリップリポジトリ
type ClipRepository interface {
	// CreateClipFolder クリップフォルダーを作成します
	//
	// 成功した場合、クリップフォルダーとnilを返します。
	// 引数に問題がある場合、ArgumentErrorを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// DBによるエラーを返すことがあります。
	CreateClipFolder(userID uuid.UUID, name, description string) (*model.ClipFolder, error)
	// UpdateClipFolder 指定したクリップフォルダーの情報を変更します
	//
	// 成功した場合、nilを返します。
	// 引数に問題がある場合、ArgumentErrorを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーを指定した場合、ErrNotFoundを返します。
	// DBによるエラーを返すことがあります。
	UpdateClipFolder(folderID uuid.UUID, name, description optional.String) error
	// DeleteClipFolder 指定したクリップフォルダーを削除します。
	//
	// 成功した場合、nilを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーを指定した場合、ErrNotFoundを返します。
	// DBによるエラーを返すことがあります。
	DeleteClipFolder(folderID uuid.UUID) error
	// DeleteClipFolderMessage 指定したクリップフォルダーのメッセージを削除します。
	//
	// 成功した場合、nilを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーメッセージを指定した場合、ErrNotFoundを返します。
	// DBによるエラーを返すことがあります。
	DeleteClipFolderMessage(folderID, messageID uuid.UUID) error
	// AddClipFolderMessage 指定したクリップフォルダーに指定したメッセージを追加します。
	//
	// 成功した場合、nilを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーを指定した場合、ErrNotFoundを返します。
	// 既に存在するフォルダーとメッセージの組み合わせを指定した場合、ErrAlreadyExistsを返します。
	// DBによるエラーを返すことがあります。
	AddClipFolderMessage(folderID, messageID uuid.UUID) (*model.ClipFolderMessage, error)
	// GetClipFoldersByUserID ユーザーのクリップフォルダーを取得します。
	//
	// 成功した場合クリップフォルダーのスライスとnilを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// DBによるエラーを返すことがあります。
	GetClipFoldersByUserID(userID uuid.UUID) ([]*model.ClipFolder, error)
	// GetClipFolder クリップフォルダーの情報を取得します。
	//
	// 成功した場合クリップフォルダーの情報とnilを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーを指定した場合、ErrNotFoundを返します。
	// DBによるエラーを返すことがあります。
	GetClipFolder(folderID uuid.UUID) (*model.ClipFolder, error)
	// GetClipFolderMessages 指定したクエリでクリップフォルダー内のメッセージのリストを取得します。
	//
	// 成功した場合クリップフォルダー内のメッセージの情報を返します。負のoffset, limitは無視されます。
	// 指定した範囲内にlimitを超えてメッセージが存在していた場合、trueを返します。
	// 引数にuuid.Nilを指定した場合、ErrNilIDを返します。
	// 存在しないクリップフォルダーを指定した場合、ErrNotFoundを返します。
	// DBによるエラーを返すことがあります。
	GetClipFolderMessages(folderID uuid.UUID, query ClipFolderMessageQuery) (messages []*model.ClipFolderMessage, more bool, err error)
	// GetMessageClips 指定したユーザーの指定したメッセージのクリップのリストを取得します。
	//
	// 成功した場合、クリップの配列とnilを返します。
	// 存在しないユーザー・メッセージを指定した場合、空配列とnilを返します。
	// DBによるエラーを返すことがあります。
	GetMessageClips(userID, messageID uuid.UUID) ([]*model.ClipFolderMessage, error)
}
