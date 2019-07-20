package bookshelf

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS library DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE library;`,
	`CREATE TABLE IF NOT EXISTS books (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		title VARCHAR(255) NULL,
		author VARCHAR(255) NULL,
		publishedDate VARCHAR(255) NULL,
		imageUrl VARCHAR(255) NULL,
		description TEXT NULL,
		createdBy VARCHAR(255) NULL,
		createdById VARCHAR(255) NULL,
		PRIMARY KEY (id)
	);`,
}

// mysqlDB persists books to a MySQL instance.
type mysqlDB struct {
	conn *sql.DB

	list   *sql.Stmt
	insert *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

// Ensure mysqlDB conforms to the BookDatabase interface.
var _ BookDatabase = &mysqlDB{}

// execSQL executes a given statement, expecting one row to be affected.
func execSQL(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}

// ListBooks lists all books, ordered by title.
func (m *mysqlDB) ListBooks() ([]*Book, error) {
	fmt.Println("DB ListBook")
	books := []*Book{
		{
			ID:     1,
			Title:  "book1",
			Author: "author1",
		},
		{
			ID:     2,
			Title:  "book2",
			Author: "author2",
		},
	}
	return books, nil
}

const getStatement = `SELECT * FROM books WHERE id = ?`

// GetBook retrieves a book by its ID.
func (m *mysqlDB) GetBook(bookid int64) (*Book, error) {
	fmt.Println("DB GetBook")
	get, err := m.conn.Prepare(getStatement)
	if err != nil {
		return nil, fmt.Errorf("mysql: prepare get: %v", err)
	}

	row := get.QueryRow(bookid)
	var (
		id            int64
		title         sql.NullString
		author        sql.NullString
		publishedDate sql.NullString
		imageUrl      sql.NullString
		description   sql.NullString
		createdBy     sql.NullString
		createdById   sql.NullString
	)
	if err := row.Scan(&id, &title, &author, &publishedDate, &imageUrl, &description, &createdBy, &createdById); err != nil {
		return nil, err
	}

	book := &Book{
		ID:     id,
		Title:  title.String,
		Author: author.String,
	}
	return book, nil
}

const insertStatement = `
INSERT INTO books (title, author) VALUES (?, ?)`

// AddBook saves a given book, assigning it a new ID
func (m *mysqlDB) AddBook(b *Book) (id int64, err error) {
	fmt.Println("DB AddBook")
	insert, err := m.conn.Prepare(insertStatement)
	if err != nil {
		return -1, fmt.Errorf("mysql: prepare insert: %v", err)
	}
	r, err := execSQL(insert, b.Title, b.Author)
	if err != nil {
		return -1, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("mysql: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

// DeleteBook removes a given book by its ID
func (m *mysqlDB) DeleteBook(id int64) error {
	fmt.Println("DB DeleteBook")
	return nil
}

// UpdateBook updates the entry for a given book
func (m *mysqlDB) UpdateBook(b *Book) error {
	fmt.Println("DB UpdateBook")
	return nil
}

// Close closes the database, freeing up resources
func (m *mysqlDB) Close() {
	fmt.Println("DB close")
}

type MySQLConfig struct {
	// Optional.
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket shoud be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the filepath to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string
}

// dataStoreName returns a connecton string suitable for sql.Open.
func (c MySQLConfig) dataStoreName(dbName string) string {
	cred := ""
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}
	if c.UnixSocket != "" {
		return fmt.Sprintf("%sunix(%s)/%s", cred, c.UnixSocket, dbName)
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, dbName)
}

func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// ensureTableExists checks the table exists. If not, it creates it.
func (c MySQLConfig) ensureTableExists() error {
	conn, err := sql.Open("mysql", c.dataStoreName(""))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE library"); err != nil {
		// MySQL error 1049 is "database does not exist".
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE books"); err != nil {
		// MySQL error 1146 is "table does not exist".
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// newMySQLDB creates a new BookDatabase backed by a given MySQL server.
func newMySQLDB(config MySQLConfig) (BookDatabase, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExists(); err != nil {
		return nil, err
	}
	conn, err := sql.Open("mysql", config.dataStoreName("library"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	db := &mysqlDB{
		conn: conn,
	}

	return db, nil
}
