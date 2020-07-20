package middleware

import (
	"bytes"
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/alt2dev/simple-wallet/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	HISTORY_DIR_DEPOSIT  = 1
	HISTORY_DIR_WITHDRAW = 2
)

var (
	createWallet = `INSERT INTO epay.wallets (balance, firstname, lastname)
		VALUES (0, $1, $2) RETURNING wallet_id`

	topupWallet1 = `UPDATE epay.wallets SET balance = balance + $1
		WHERE wallet_id = $2`

	topupWallet2 = `INSERT INTO epay.transactions
		(from_id, to_id, value, date)
		VALUES (NULL, $1, $2, current_timestamp)
		RETURNING transaction_id`

	sendToWallet1 = `UPDATE epay.wallets
		SET balance = balance - $1
		WHERE wallet_id = $2`

	sendToWallet2 = `UPDATE epay.wallets SET balance = balance + $1
		WHERE wallet_id = $2`

	sendToWallet3 = `INSERT INTO epay.transactions
		(from_id, to_id, value, date)
		VALUES ($1, $2, $3, current_timestamp)
		RETURNING transaction_id`

	historyOfWalletD = `SELECT * FROM epay.transactions WHERE
		to_id = $1 AND DATE(date) = DATE($2)`

	historyOfWalletW = `SELECT * FROM epay.transactions WHERE
		from_id = $1 AND DATE(date) = DATE($2)`
)

var (
	pg = database{}
)

type database struct {
	*pgxpool.Pool
	*pgxpool.Config
}

func ParseDB(dbURL string) error {
	return pg.parse(dbURL)
}

func InitDB() error {
	return pg.init()
}

func CloseDB() {
	pg.Pool.Close()
}

func CreatePOST(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if wallet.Firstname == "" || wallet.Lastname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "no empty string"})
		return
	}

	walletId, err := pg.createWallet(wallet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": walletId,
	})
}

func TopupPOST(c *gin.Context) {
	var topup models.Topup
	if err := c.ShouldBindJSON(&topup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transactionId, err := pg.topupWallet(topup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"transaction-id": transactionId,
	})
}

func SendPOST(c *gin.Context) {
	var send models.Send
	if err := c.ShouldBindJSON(&send); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	transactionId, err := pg.sendToWallet(send)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transaction_id": transactionId,
	})
}

func HistoryGET(c *gin.Context) {
	var dir int
	var data [][]string

	walletId := c.Param("id")
	if ok := validWalletId(walletId, 32); !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   "error",
			"parametr": "id",
			"msg":      "invalid wallet id",
		})
		return
	}

	date, ok := validDate(c.Query("date"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   "error",
			"parametr": "date",
			"msg":      "invalid date format (dd-mm-yyyy)",
		})
		return
	}

	direction := c.Query("direction")
	switch direction {
	case "deposit":
		dir = HISTORY_DIR_DEPOSIT
	case "withdraw":
		dir = HISTORY_DIR_WITHDRAW
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":   "error",
			"parametr": "direction",
			"msg":      "invalid parametr (deposit|withdraw)",
		})
		return
	}

	transactions, err := pg.historyOfWallet(dir, walletId, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, t := range transactions {
		data = append(data, t.ToSliceOfStrings())
	}

	byteBuf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(byteBuf)

	csvWriter.WriteAll(data)

	if err := csvWriter.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, byteBuf.String())
}

func (pg *database) parse(dbURL string) error {
	var err error
	pg.Config, err = pgxpool.ParseConfig(dbURL)
	return err
}

func (pg *database) init() error {
	var err error
	pg.Pool, err = pgxpool.ConnectConfig(context.Background(), pg.Config)
	return err
}

func (pg *database) createWallet(wallet models.Wallet) (string, error) {
	var newWalletId string
	err := pg.Pool.QueryRow(context.Background(), createWallet, wallet.Firstname, wallet.Lastname).Scan(&newWalletId)
	return newWalletId, err
}

func (pg *database) topupWallet(topup models.Topup) (int64, error) {
	tx, err := pg.Pool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	tx.Exec(context.Background(), "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")

	_, err = tx.Exec(context.Background(), topupWallet1, topup.Amount, topup.RecipientId)
	if err != nil {
		return 0, err
	}

	var transactionId int64
	row := tx.QueryRow(context.Background(), topupWallet2, topup.RecipientId, topup.Amount)
	row.Scan(&transactionId)

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}
	return transactionId, nil
}

func (pg *database) sendToWallet(send models.Send) (int64, error) {
	tx, err := pg.Pool.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	tx.Exec(context.Background(), "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")

	_, err = tx.Exec(context.Background(), sendToWallet1, send.Amount, send.SenderId)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(context.Background(), sendToWallet2, send.Amount, send.RecipientId)
	if err != nil {
		return 0, err
	}

	var transactionId int64
	row := tx.QueryRow(context.Background(), sendToWallet3, send.SenderId, send.RecipientId, send.Amount)
	row.Scan(&transactionId)

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}
	return transactionId, nil
}

func (pg *database) historyOfWallet(dir int, walletId string, date time.Time) ([]models.Transaction, error) {
	var tmpSQL string
	var ts []models.Transaction
	var tmpSenderId pgtype.Varchar

	tx, err := pg.Pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	tx.Exec(context.Background(), "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")

	switch dir {
	case HISTORY_DIR_DEPOSIT:
		tmpSQL = historyOfWalletD
	case HISTORY_DIR_WITHDRAW:
		tmpSQL = historyOfWalletW
	}

	rows, err := tx.Query(context.Background(), tmpSQL, walletId, date)
	defer rows.Close()

	for rows.Next() {
		var t models.Transaction
		err = rows.Scan(&t.TransactionId, &tmpSenderId, &t.RecipientId, &t.Date, &t.Amount)
		if err != nil {
			return nil, err
		}

		switch tmpSenderId.Status {
		case pgtype.Null:
			t.SenderId = ""
		case pgtype.Present:
			t.SenderId = tmpSenderId.String
		default:
			t.SenderId = ""
		}

		ts = append(ts, t)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}
	return ts, nil
}

func validWalletId(walletId string, size int) bool {
	for _, r := range walletId {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	if utf8.RuneCountInString(walletId) != size {
		return false
	}
	return true
}

func validAmount(amount string) (uint64, string, bool) {
	msg := ``
	u, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return 0, msg, false
	}
	return u, "", true
}

func validDate(inputDate string) (time.Time, bool) {
	const layout = "02-01-2006"
	date, err := time.Parse(layout, inputDate)
	if err != nil {
		return time.Time{}, false
	}
	return date, true
}
