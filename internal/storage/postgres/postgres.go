package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/ST359/avito-trainee-backend-winter-2025/internal/storage"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}
type UserInfo struct {
	CoinHistory CoinHistory
	Coins       int
	Inventory   []InventoryEntry
}
type CoinHistory struct {
	Received []TransactionReceived
	Sent     []TransactionSent
}
type TransactionReceived struct {
	Amount   int
	FromUser string
}
type TransactionSent struct {
	Amount int
	ToUser string
}
type InventoryEntry struct {
	Quantity int
	Type     string
}

func New(port, user, password, name, host string) (*Storage, error) {
	const op = "storage.postgres.New"
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil

}

func (s *Storage) AddUser(name, passHash string) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_, err := psql.Insert("users").
		Columns("name", "pass_hash").
		Values(name, passHash).
		RunWith(s.db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) UserPassHash(name string) (string, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var passHash string
	err := psql.Select("pass_hash").From("users").Where("name=?", name).
		RunWith(s.db).Scan(&passHash)
	if err != nil {
		return "", err
	}
	return passHash, nil
}

func (s *Storage) UserExist(name string) (bool, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var count int
	err := psql.Select("COUNT(*)").From("users").Where("name=?", name).RunWith(s.db).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (s *Storage) SendCoins(fromUser string, toUser string, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	var fromCoins, toCoins int
	var fromUserId, toUserId int
	err = psql.Select("id", "coins").
		From("users").
		Where("name=?", fromUser).
		RunWith(tx).
		QueryRow().
		Scan(&fromUserId, &fromCoins)
	if err != nil {
		return fmt.Errorf("failed to get coins for fromUser: %w", err)
	}
	if fromCoins < amount {
		return storage.ErrUnsufficientBalance
	}
	err = psql.Select("id", "coins").
		From("users").
		Where("name=?", toUser).
		RunWith(tx).
		QueryRow().
		Scan(&toUserId, &toCoins)
	if err != nil {
		return fmt.Errorf("failed to get coins for toUser: %w", err)
	}

	if fromCoins < amount {
		return storage.ErrUnsufficientBalance
	}

	_, err = psql.Update("users").
		Set("coins", fromCoins-amount).
		Where("name=?", fromUser).
		RunWith(tx).
		Exec()
	if err != nil {
		return fmt.Errorf("failed to update coins for fromUser: %w", err)
	}

	_, err = psql.Update("users").
		Set("coins", toCoins+amount).
		Where("name=?", toUser).
		RunWith(tx).
		Exec()
	if err != nil {
		return fmt.Errorf("failed to update coins for toUser: %w", err)
	}
	_, err = psql.Insert("transactions").
		Columns("from_user_id", "to_user_id", "amount").
		Values(fromUserId, toUserId, amount).
		RunWith(tx).
		Exec()
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (s *Storage) UserInfo(user string) (*UserInfo, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var userInfo UserInfo
	var userID int
	//balance
	err := psql.Select("id", "coins").
		From("users").
		Where("name=?", user).
		RunWith(s.db).
		QueryRow().
		Scan(&userID, &userInfo.Coins)
	if err != nil {
		return nil, err
	}
	//inventory
	rows, err := psql.Select("m.name", "ui.quantity").
		From("user_inventory ui").
		Join("merch m ON ui.merch_id = m.id").
		Where("ui.user_id = ?", userID).
		RunWith(s.db).
		Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ie InventoryEntry
		if err := rows.Scan(&ie.Type, &ie.Quantity); err != nil {
			return nil, err
		}
		userInfo.Inventory = append(userInfo.Inventory, ie)
	}
	//coin history
	//transactions SENT
	tsRows, err := psql.Select("u.name", "t.amount").
		From("transactions t").
		Join("users u ON u.id = t.to_user_id").
		Where("t.from_user_id = ?", userID).
		RunWith(s.db).
		Query()
	if err != nil {
		return nil, err
	}
	for tsRows.Next() {
		var (
			trSent TransactionSent
		)
		if err := tsRows.Scan(&trSent.ToUser, &trSent.Amount); err != nil {
			return nil, err
		}
		userInfo.CoinHistory.Sent = append(userInfo.CoinHistory.Sent, trSent)
	}
	//transactions RECEIVED
	trRows, err := psql.Select("u.name", "t.amount").
		From("transactions t").
		Join("users u ON u.id = t.from_user_id").
		Where("t.to_user_id = ?", userID).
		RunWith(s.db).
		Query()
	if err != nil {
		return nil, err
	}
	for trRows.Next() {
		var (
			trRcv TransactionReceived
		)
		if err := trRows.Scan(&trRcv.FromUser, &trRcv.Amount); err != nil {
			return nil, err
		}
		userInfo.CoinHistory.Received = append(userInfo.CoinHistory.Received, trRcv)
	}
	return &userInfo, nil
}

func (s *Storage) Buy(item string, user string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var userBalance, userID, itemPrice, itemID int
	err = psql.Select("id", "coins").
		From("users").
		Where("name=?", user).
		RunWith(tx).
		QueryRow().
		Scan(&userID, &userBalance)
	if err != nil {
		return fmt.Errorf("failed to get coins for user: %w", err)
	}

	err = psql.Select("price", "id").
		From("merch").
		Where("name=?", item).
		RunWith(tx).
		QueryRow().
		Scan(&itemPrice, &itemID)
	if err != nil {
		return fmt.Errorf("failed to get items info: %w", err)
	}

	if userBalance < itemPrice {
		return storage.ErrUnsufficientBalance
	}

	_, err = psql.Update("users").
		Set("coins", userBalance-itemPrice).
		Where("name=?", user).
		RunWith(tx).
		Exec()
	if err != nil {
		return fmt.Errorf("failed to update coins for fromUser: %w", err)
	}

	_, err = psql.Insert("user_inventory").
		Columns("user_id", "merch_id", "quantity").
		Values(userID, itemID, 1).
		RunWith(tx).
		Suffix("ON CONFLICT (user_id, merch_id) DO UPDATE SET quantity = user_inventory.quantity + EXCLUDED.quantity").
		Exec()
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) ItemExist(name string) (bool, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var count int
	err := psql.Select("COUNT(*)").From("merch").Where("name=?", name).RunWith(s.db).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
