package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type PortalRepository struct {
	*sql.DB
}

func NewPostgresDBRepository() (*PortalRepository, error) {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
		log.Println("POSTGRES_HOST no definido, usando 'localhost'")
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
		log.Println("POSTGRES_PORT no definido, usando '5432'")
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
		log.Println("POSTGRES_USER no definido, usando 'postgres'")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
		log.Println("POSTGRES_PASSWORD no definido, usando valor por defecto")
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname == "" {
		dbname = "wemaps"
		log.Println("POSTGRES_DB no definido, usando 'wemaps'")
	}

	return makePostgresDB(host, port, user, password, dbname)
}

func makePostgresDB(host, port, user, password, dbname string) (*PortalRepository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	log.Println("Conexión a PostgreSQL establecida")
	return &PortalRepository{db}, nil
}

func (db *PortalRepository) Close() error {
	return db.DB.Close()
}

func (db *PortalRepository) GetUserID(alias string) (int, error) {
	var userID int
	query := "SELECT id FROM users WHERE alias = $1"
	err := db.QueryRow(query, alias).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("error querying user ID: %v", err)
	}
	return userID, nil
}

func (db *PortalRepository) CreateUser(email, alias, fullName, phone string) (int, error) {
	var userID int
	query := `INSERT INTO users (email, alias, full_name, phone)
			  VALUES ($1, $2, $3, $4)
			  RETURNING id`
	err := db.QueryRow(query, email, alias, fullName, phone).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("error creating user: %v", err)
	}
	return userID, nil
}

/*

//create method to create the users table in PostgreSQL
-- SQL script to create the users table in PostgreSQL

CREATE TABLE public.users (
	id serial4 NOT NULL,
	email varchar(255) NOT NULL,
	alias varchar(255) NULL,
	full_name varchar(255) NULL,
	phone varchar(20) NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_users_email ON public.users USING btree (email);

-- Permissions

ALTER TABLE public.users OWNER TO root;
GRANT ALL ON TABLE public.users TO root;
*/
