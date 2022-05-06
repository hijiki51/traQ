package twemoji

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/repository"
	"github.com/traPtitech/traQ/service/file"
	"github.com/traPtitech/traQ/utils/optional"
)

const (
	/*
		twemoji Copyright 2019 Twitter, Inc and other contributors
		Graphics licensed under CC-BY 4.0: https://creativecommons.org/licenses/by/4.0/
	*/
	emojiZipURL  = "https://github.com/twitter/twemoji/archive/v14.0.2.zip"
	emojiDir     = "twemoji-14.0.2/assets/svg/"
	emojiMetaURL = "https://raw.githubusercontent.com/joypixels/emoji-toolkit/master/emoji.json"
)

type emojiMeta struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Order      int    `json:"order"`
	ShortName  string `json:"shortname"`
	CodePoints struct {
		FullyQualified string   `json:"fully_qualified"`
		DefaultMatches []string `json:"default_matches"`
	} `json:"code_points"`
}

func Install(repo repository.Repository, fm file.Manager, logger *zap.Logger, update bool) error {
	logger = logger.Named("twemoji_installer")

	// 絵文字メタデータをダウンロード
	logger.Info("downloading meta data...: " + emojiMetaURL)
	emojis, err := downloadEmojiMeta()
	if err != nil {
		return err
	}
	logger.Info("finished downloading meta data")

	// 絵文字画像データをダウンロード
	logger.Info("downloading twemoji...: " + emojiZipURL)
	twemojiZip, err := downloadEmojiZip()
	if err != nil {
		return err
	}
	logger.Info("finished downloading twemoji")

	// 絵文字解凍・インストール
	zipfile, err := zip.NewReader(twemojiZip, twemojiZip.Size())
	if err != nil {
		return err
	}

	saveEmojiFile := func(f *zip.File) (model.File, error) {
		_, filename := path.Split(f.Name)
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()

		return fm.Save(file.SaveArgs{
			FileName: filename,
			FileSize: f.FileInfo().Size(),
			FileType: model.FileTypeStamp,
			Src:      r,
		})
	}

	logger.Info("installing emojis...")
	for _, file := range zipfile.File {
		if file.FileInfo().IsDir() {
			continue
		}

		dir, filename := path.Split(file.Name)
		if dir != emojiDir || !strings.HasSuffix(filename, ".svg") {
			continue
		}

		code := strings.TrimSuffix(filename, ".svg")
		emoji, ok := emojis[code]
		if !ok {
			emoji, ok = emojis["00"+code]
			if !ok {
				continue
			}
		}

		name := strings.Trim(emoji.ShortName, ":")
		s, err := repo.GetStampByName(name)
		if err != nil && err != repository.ErrNotFound {
			return err
		}

		if s == nil {
			// 新規追加
			meta, err := saveEmojiFile(file)
			if err != nil {
				return err
			}

			s, err := repo.CreateStamp(repository.CreateStampArgs{
				Name:      name,
				FileID:    meta.GetID(),
				CreatorID: uuid.Nil,
				IsUnicode: true,
			})
			if err != nil {
				return err
			}

			logger.Info(fmt.Sprintf("stamp added: %s (%s)", name, s.ID))
		} else {
			if !update {
				continue
			}

			// 既存のファイルを置き換え
			meta, err := saveEmojiFile(file)
			if err != nil {
				return err
			}

			if err := repo.UpdateStamp(s.ID, repository.UpdateStampArgs{
				FileID: optional.UUIDFrom(meta.GetID()),
			}); err != nil {
				return err
			}

			logger.Info(fmt.Sprintf("stamp updated: %s (%s)", name, s.ID))
		}
	}
	logger.Info("finished installing emojis")
	return nil
}

func downloadEmojiZip() (*bytes.Reader, error) {
	res, err := http.Get(emojiZipURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	return bytes.NewReader(b), err
}

func downloadEmojiMeta() (map[string]*emojiMeta, error) {
	res, err := http.Get(emojiMetaURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var temp map[string]*emojiMeta
	if err := jsoniter.ConfigFastest.NewDecoder(res.Body).Decode(&temp); err != nil {
		return nil, err
	}

	emojis := map[string]*emojiMeta{}
	for _, v := range temp {
		if v.Category == "modifier" {
			continue
		}
		if strings.HasSuffix(v.Name, "skin tone") {
			continue
		}

		emojis[v.CodePoints.FullyQualified] = v
		for _, s := range v.CodePoints.DefaultMatches {
			emojis[s] = v
		}
	}
	return emojis, nil
}
