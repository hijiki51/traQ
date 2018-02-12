package model

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDevice_TableName(t *testing.T) {
	assert.Equal(t, "devices", (&Device{}).TableName())
}

func TestDevice_Register(t *testing.T) {
	beforeTest(t)
	assert := assert.New(t)

	id1 := testUserID
	id2 := privateUserID
	token1 := "ajopejiajgopnavdnva8y48fhaerudsyf8uf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token2 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"

	assert.NoError((&Device{UserId: id1, Token: token1}).Register())
	assert.NoError((&Device{UserId: id2, Token: token2}).Register())
	assert.Error((&Device{UserId: id1, Token: token2}).Register())

	l, _ := db.Count(&Device{})
	assert.Equal(int64(2), l)
}

func TestDevice_Unregister(t *testing.T) {
	beforeTest(t)
	assert := assert.New(t)

	id1 := testUserID
	id2 := privateUserID
	token1 := "ajopejiajgopnavdnva8y48fhaerudsyf8uf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token2 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token3 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvfawfefwfwe3iqodwjkdvlznfjbxdefpuw90jiosdv"

	assert.NoError((&Device{UserId: id1, Token: token1}).Register())
	assert.NoError((&Device{UserId: id2, Token: token2}).Register())
	assert.NoError((&Device{UserId: id1, Token: token3}).Register())

	assert.NoError((&Device{Token: token2}).Unregister())
	l, _ := db.Count(&Device{})
	assert.Equal(int64(2), l)

	assert.NoError((&Device{UserId: id1}).Unregister())
	l, _ = db.Count(&Device{})
	assert.Equal(int64(0), l)
}

func TestGetAllDevices(t *testing.T) {
	beforeTest(t)
	assert := assert.New(t)

	id1 := testUserID
	id2 := privateUserID
	token1 := "ajopejiajgopnavdnva8y48fhaerudsyf8uf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token2 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token3 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvfawfefwfwe3iqodwjkdvlznfjbxdefpuw90jiosdv"

	assert.NoError((&Device{UserId: id1, Token: token1}).Register())
	assert.NoError((&Device{UserId: id2, Token: token2}).Register())
	assert.NoError((&Device{UserId: id1, Token: token3}).Register())

	devs, err := GetAllDevices()
	assert.NoError(err)
	assert.Len(devs, 3)
}

func TestGetDevices(t *testing.T) {
	beforeTest(t)
	assert := assert.New(t)

	id1 := testUserID
	id2 := privateUserID
	token1 := "ajopejiajgopnavdnva8y48fhaerudsyf8uf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token2 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvlfjxhgyru83iqodwjkdvlznfjbxdefpuw90jiosdv"
	token3 := "ajopejiajgopnavdnva8y48ffwefwefewfwf39ifoewkvfawfefwfwe3iqodwjkdvlznfjbxdefpuw90jiosdv"

	assert.NoError((&Device{UserId: id1, Token: token1}).Register())
	assert.NoError((&Device{UserId: id2, Token: token2}).Register())
	assert.NoError((&Device{UserId: id1, Token: token3}).Register())

	devs, err := GetDevices(uuid.FromStringOrNil(id1))
	assert.NoError(err)
	assert.Len(devs, 2)

	devs, err = GetDevices(uuid.FromStringOrNil(id2))
	assert.NoError(err)
	assert.Len(devs, 1)
}
