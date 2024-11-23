package database

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testDbPath = "./testDb.db"
)

func dropDatabase(fileName string) {
	_ = os.Remove(fileName)
}

func clearDb() {
	dropDatabase(testDbPath)
}

func connectDb(t *testing.T) *OneTimeShareDb {
	assert := require.New(t)
	db, err := ConnectDb(testDbPath)

	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return nil
	}
	return db
}

func createDbAndConnect(t *testing.T) *OneTimeShareDb {
	clearDb()
	return connectDb(t)
}

func TestConnection(t *testing.T) {
	assert := require.New(t)
	dropDatabase(testDbPath)

	db, err := ConnectDb(testDbPath)

	defer dropDatabase(testDbPath)
	if err != nil {
		assert.Fail("Problem with creation db connection:" + err.Error())
		return
	}

	assert.True(db.IsConnectionOpened())

	db.Disconnect()

	assert.False(db.IsConnectionOpened())
}

func TestSanitizeString(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	testText := "text'test''test\"test\\"

	db.SetDatabaseVersion(testText)
	assert.Equal(testText, db.GetDatabaseVersion())
}

func TestDatabaseVersion(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}

	{
		version := db.GetDatabaseVersion()
		assert.Equal(latestVersion, version)
	}

	{
		db.SetDatabaseVersion("1.0")
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
	}

	db.Disconnect()

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.0", version)
		db.Disconnect()
	}

	{
		db = connectDb(t)
		db.SetDatabaseVersion("1.2")
		db.Disconnect()
	}

	{
		db = connectDb(t)
		version := db.GetDatabaseVersion()
		assert.Equal("1.2", version)
		db.Disconnect()
	}
}

func TestGetUserLimits(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var token1 = "321"
	var token2 = "123"

	{
		isFound, _, _, _ := db.GetUserLimits(token1)
		assert.False(isFound)
	}

	assert.False(db.DoesUserExist(token1))

	db.SetUserLimits(token1, 1, 2, 3)
	assert.True(db.DoesUserExist(token1))

	{
		isFound, retentionLimitMinutes, maxSizeBytes, shareCreationLimitMinutes := db.GetUserLimits(token1)

		assert.True(isFound)
		assert.False(db.DoesUserExist(token2))
		assert.Equal(1, retentionLimitMinutes)
		assert.Equal(2, maxSizeBytes)
		assert.Equal(3, shareCreationLimitMinutes)
	}
}

func TestRemoveUserLimits(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	var token1 = "321"
	var token2 = "123"

	db.SetUserLimits(token1, 1, 2, 3)
	assert.True(db.DoesUserExist(token1))

	db.RemoveUserByToken(token1)
	assert.False(db.DoesUserExist(token1))

	db.RemoveUserByToken(token2)
	assert.False(db.DoesUserExist(token2))
}

func TestUpdateUserLimits(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	token := "321"

	db.SetUserLimits(token, 1, 2, 3)

	{
		isFound, retentionLimitMinutes, maxSizeBytes, shareCreationLimitMinutes := db.GetUserLimits(token)
		assert.True(isFound)
		assert.Equal(1, retentionLimitMinutes)
		assert.Equal(2, maxSizeBytes)
		assert.Equal(3, shareCreationLimitMinutes)
	}

	db.SetUserLimits(token, 4, 5, 6)

	{
		isFound, retentionLimitMinutes, maxSizeBytes, shareCreationLimitMinutes := db.GetUserLimits(token)
		assert.True(isFound)
		assert.Equal(4, retentionLimitMinutes)
		assert.Equal(5, maxSizeBytes)
		assert.Equal(6, shareCreationLimitMinutes)
	}
}

func TestSaveAndConsumeMessage(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	const formatVersion = 0

	var message1 = "test message 1"
	var message2 = "test message 2"
	var message3 = "test message 3"

	var messageToken1 = "321"
	var messageToken2 = "123"

	err := db.SaveMessage(messageToken1, 100, message1, formatVersion)
	assert.Nil(err)
	err = db.SaveMessage(messageToken1, 200, message2, formatVersion)
	assert.NotNil(err)
	err = db.SaveMessage(messageToken2, 300, message3, formatVersion)
	assert.Nil(err)

	{
		message, expireTimestamp, _ := db.TryConsumeMessage(messageToken1)
		assert.Equal(message1, *message)
		assert.Equal(int64(100), expireTimestamp)
	}

	{
		message, _, _ := db.TryConsumeMessage(messageToken1)
		assert.Nil(message)
	}

	{
		message, expireTimestamp, _ := db.TryConsumeMessage(messageToken2)
		assert.Equal(message3, *message)
		assert.Equal(int64(300), expireTimestamp)
	}

	{
		message, _, _ := db.TryConsumeMessage(messageToken2)
		assert.Nil(message)
	}

	{
		message, _, _ := db.TryConsumeMessage("not existing token")
		assert.Nil(message)
	}
}

func TestClearExpiredMessages(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	const formatVersion = 0

	var message1 = "test message 1"
	var message2 = "test message 2"

	var messageToken1 = "321"
	var messageToken2 = "123"

	err := db.SaveMessage(messageToken1, 100, message1, formatVersion)
	assert.Nil(err)
	err = db.SaveMessage(messageToken2, 200, message2, formatVersion)
	assert.Nil(err)

	db.ClearExpiredMessages(160)

	{
		message, _, _ := db.TryConsumeMessage(messageToken1)
		assert.Nil(message)
	}

	{
		message, expireTimestamp, _ := db.TryConsumeMessage(messageToken2)
		assert.Equal(message2, *message)
		assert.Equal(int64(200), expireTimestamp)
	}
}

func TestUserLastMessageCreationTime(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	token := "123"

	db.SetUserLimits(token, 1, 2, 3)

	{
		lastTime := db.GetUserLastMessageCreationTime(token)
		assert.Equal(int64(0), lastTime)
	}

	{
		db.SetUserLastMessageCreationTime(token, 100)
		lastTime := db.GetUserLastMessageCreationTime(token)
		assert.Equal(int64(100), lastTime)
	}

	{
		db.SetUserLastMessageCreationTime(token, 200)
		lastTime := db.GetUserLastMessageCreationTime(token)
		assert.Equal(int64(200), lastTime)
	}
}

func TestSettingLimitsDoesNotChangeLastMessageCreationTime(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	token := "123"
	db.SetUserLimits(token, 0, 0, 0)

	db.SetUserLastMessageCreationTime(token, 100)
	assert.Equal(int64(100), db.GetUserLastMessageCreationTime(token))

	db.SetUserLimits(token, 1, 2, 3)

	assert.Equal(int64(100), db.GetUserLastMessageCreationTime(token))
}

func TestSaveMessageWithVersionConsumedMessageRetainVersion(t *testing.T) {
	assert := require.New(t)
	db := createDbAndConnect(t)
	defer clearDb()
	if db == nil {
		t.Fail()
		return
	}
	defer db.Disconnect()

	const oldFormatVersion = uint32(0)
	const newFormatVersion = uint32(1)

	var message1 = "test message"
	var messageToken1 = "123"
	var message2 = "test message 2"
	var messageToken2 = "321"

	err := db.SaveMessage(messageToken1, 100, message1, oldFormatVersion)
	assert.Nil(err)
	err = db.SaveMessage(messageToken2, 200, message2, newFormatVersion)
	assert.Nil(err)

	{
		message, _, formatVersion := db.TryConsumeMessage(messageToken1)
		assert.Equal(message1, *message)
		assert.Equal(oldFormatVersion, formatVersion)
	}

	{
		message, _, formatVersion := db.TryConsumeMessage(messageToken2)
		assert.Equal(message2, *message)
		assert.Equal(newFormatVersion, formatVersion)
	}
}
