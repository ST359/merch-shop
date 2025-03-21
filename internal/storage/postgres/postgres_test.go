package postgres

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestUserExist(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	storage := &Storage{db: db}
	//testing existing user
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE name=$1").
		WithArgs("existinguser").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := storage.UserExist("existinguser")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())

	//testing non-existing user
	mock.ExpectQuery("SELECT COUNT(*) FROM users WHERE name=$1").
		WithArgs("nonexistinguser").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err = storage.UserExist("nonexistinguser")
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestAddUser(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	storage := &Storage{db: db}

	mock.ExpectExec("INSERT INTO users (name,pass_hash) VALUES ($1,$2)").
		WithArgs("testuser", "hashedpassword").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.AddUser("testuser", "hashedpassword")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserPassHash(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	storage := &Storage{db: db}

	mock.ExpectQuery("SELECT pass_hash FROM users WHERE name=$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"pass_hash"}).AddRow("hashedpassword"))

	passHash, err := storage.UserPassHash("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "hashedpassword", passHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCoinsSuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	fromUser := "fromUser"
	toUser := "toUser"
	amount := 10
	initBalance := 1000
	//Expecting that amount will be substracted from fromUser balance and added to toUser balance
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(fromUser).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, initBalance)) // init balance fromUser
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(toUser).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(2, initBalance)) // init balance toUser
	mock.ExpectExec("UPDATE users SET coins = $1 WHERE name=$2").
		WithArgs(initBalance-amount, fromUser).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE users SET coins = $1 WHERE name=$2").
		WithArgs(initBalance+amount, toUser).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO transactions (from_user_id,to_user_id,amount) VALUES ($1,$2,$3)").
		WithArgs(1, 2, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = s.SendCoins(fromUser, toUser, amount)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
func TestSendCoinsInsufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	fromUser := "fromUser"
	toUser := "toUser"
	amount := 1010
	initBalance := 1000
	//Expecting that insufficient balance error will return
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(fromUser).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, initBalance))
	mock.ExpectRollback()

	err = s.SendCoins(fromUser, toUser, amount)

	assert.ErrorIs(t, err, storage.ErrUnsufficientBalance)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
func TestSendCoinsUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	fromUser := "fromUser"
	toUser := "toUser"
	amount := 100

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(fromUser).
		WillReturnError(fmt.Errorf("user not found"))
	mock.ExpectRollback()

	err = s.SendCoins(fromUser, toUser, amount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get coins for fromUser")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUserInfo(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	storage := &Storage{db: db}
	coins := 100
	//Coins
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, coins))

	//Inventory
	mock.ExpectQuery("SELECT m.name, ui.quantity FROM user_inventory ui JOIN merch m ON ui.merch_id = m.id WHERE ui.user_id = $1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "quantity"}).
			AddRow("item1", 2).
			AddRow("item2", 3))

	//Transactions sent
	mock.ExpectQuery("SELECT u.name, t.amount FROM transactions t JOIN users u ON u.id = t.to_user_id WHERE t.from_user_id = $1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "amount"}).AddRow("to1", 5).AddRow("to2", 10))

	//Transactions received
	mock.ExpectQuery("SELECT u.name, t.amount FROM transactions t JOIN users u ON u.id = t.from_user_id WHERE t.to_user_id = $1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "amount"}).AddRow("from1", 5).AddRow("from2", 10))
	userInfo, err := storage.UserInfo("testuser")
	if err != nil {
		t.Errorf("error was not expected while getting user info: %s", err)
	}

	// Assertions
	assert.NotNil(t, userInfo)
	assert.Equal(t, coins, userInfo.Coins)
	assert.Len(t, userInfo.Inventory, 2)
	assert.Equal(t, "item1", userInfo.Inventory[0].Type)
	assert.Equal(t, 2, userInfo.Inventory[0].Quantity)
	assert.Len(t, userInfo.CoinHistory.Sent, 2)
	assert.Equal(t, "to1", userInfo.CoinHistory.Sent[0].ToUser)
	assert.Equal(t, 5, userInfo.CoinHistory.Sent[0].Amount)
	assert.Len(t, userInfo.CoinHistory.Received, 2)
	assert.Equal(t, "from1", userInfo.CoinHistory.Received[0].FromUser)
	assert.Equal(t, 5, userInfo.CoinHistory.Received[0].Amount)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestBuySuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	buyer := "buyer"
	item := "t-shirt"
	initBalance := 1000
	price := 100
	//Expecting that amount will be substracted from buyer balance and added to toUser balance
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(buyer).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, initBalance)) // init balance for buyer
	mock.ExpectQuery("SELECT price, id FROM merch WHERE name=$1").
		WithArgs(item).
		WillReturnRows(sqlmock.NewRows([]string{"price", "id"}).AddRow(price, 1)) // merch price
	mock.ExpectExec("UPDATE users SET coins = $1 WHERE name=$2").
		WithArgs(initBalance-price, buyer).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO user_inventory (user_id,merch_id,quantity) VALUES ($1,$2,$3) ON CONFLICT (user_id, merch_id) DO UPDATE SET quantity = user_inventory.quantity + EXCLUDED.quantity").
		WithArgs(1, 1, 1).
		WillReturnResult(sqlmock.NewResult(1, 3))
	mock.ExpectCommit()

	err = s.Buy(item, buyer)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
func TestBuyInsufficientBalance(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer db.Close()

	s := &Storage{db: db}

	buyer := "buyer"
	item := "t-shirt"
	initBalance := 100
	price := 1000
	//Expecting that insufficient balance error will return
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, coins FROM users WHERE name=$1").
		WithArgs(buyer).
		WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(1, initBalance)) // init balance for buyer
	mock.ExpectQuery("SELECT price, id FROM merch WHERE name=$1").
		WithArgs(item).
		WillReturnRows(sqlmock.NewRows([]string{"price", "id"}).AddRow(price, 1)) // merch price
	mock.ExpectRollback()

	err = s.Buy(item, buyer)

	assert.ErrorIs(t, err, storage.ErrUnsufficientBalance)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
