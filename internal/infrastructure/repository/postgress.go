package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"wemaps/internal/adapters/http/dto"

	_ "github.com/lib/pq"
)

type User struct {
    ID       int    `json:"id"`
    Email    string `json:"email"`
    Alias    string `json:"alias"`
    FullName string `json:"full_name"`
    Phone    string `json:"phone"`
}

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

func (db *PortalRepository) FindUserByToken(token string) (*User, error) {
    query := `
        SELECT u.id, u.email, u.alias, u.full_name, u.phone
        FROM sessions s
        JOIN users u ON s.user_id = u.id
        WHERE s.token = $1 AND s.is_active = true AND s.expires_at > CURRENT_TIMESTAMP
    `
    var user User
    err := db.QueryRow(query, token).Scan(
        &user.ID,
        &user.Email,
        &user.Alias,
        &user.FullName,
        &user.Phone,
    )
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("no active session found for token")
    }
    if err != nil {
        log.Printf("error finding user by token: %v", err)
        return nil, fmt.Errorf("error querying session: %v", err)
    }
    return &user, nil
}

func (db *PortalRepository) GetUserID(alias string) (int, error) {
	var userID int
	query := "SELECT id FROM users WHERE alias = $1"
	err := db.QueryRow(query, alias).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, fmt.Errorf("user not found")
		}
		return -1, fmt.Errorf("error querying user ID: %v", err)
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

func (db *PortalRepository) LogSession(sessionID string, userID int, tokenString string, ipAddress string, expiresAt time.Time, active bool) {
	query := `INSERT INTO sessions (session_id, user_id, token, ip_address, expires_at, is_active)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, sessionID, userID, tokenString, ipAddress, expiresAt, active)
	if err != nil {
		log.Printf("error logging session: %v", err)
		return
	}
	log.Println("Session logged successfully")
}

func (db *PortalRepository) SaveAddress(idReport int, address string, latitude float64, longitude float64, formatAddress string, geocoder string) (int, error) {
	var addressID int
	queryCheck := `SELECT id,address,normalized_address FROM address WHERE address = $1`
	addressDB := ""
	formatAddressDB := ""
	err := db.QueryRow(queryCheck, address).Scan(&addressID, &addressDB, &formatAddressDB)
	if err == nil {
		//Cuando hay un resultado de latitud y longitud, se actualiza la dirección
		if (addressDB != address || formatAddressDB != formatAddress) {
			updateQuery := `UPDATE address SET address = $1, normalized_address = $2, latitude = $3, longitude = $4, geocoder = $5 WHERE id = $6`
			_, err = db.Exec(updateQuery, address, formatAddress, latitude, longitude, geocoder, addressID);
		}else{
			_,err = db.SaveAddressInReport(idReport, addressID, latitude, longitude, formatAddress, geocoder)
			if err != nil {
				return addressID, fmt.Errorf("error linking existing address to report: %v", err)
			}
		}
		return addressID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error checking existing address: %v", err)
	}

	queryInsert := `INSERT INTO address (address, normalized_address, latitude, longitude, geocoder)
				   VALUES ($1, $2, $3, $4, $5)
				   RETURNING id`

	err = db.QueryRow(queryInsert, address, formatAddress, latitude, longitude, geocoder).Scan(&addressID)
	if err != nil {
		log.Printf("error saving address: %v", err)
		return 0, fmt.Errorf("error saving address: %v", err)
	}	
	return addressID, nil
}

func (db *PortalRepository) SaveAddressInReport(idReport int, addressID int, latitude float64, longitude float64, formatAddress string, geocoder string) (int, error) {

	linkQuery := `INSERT INTO report_address (report_id, address_id)
				  VALUES ($1, $2)`
	_, err := db.Exec(linkQuery, idReport, addressID)
	if err != nil {
		log.Printf("error linking address to report: %v", err)
		return addressID, fmt.Errorf("error linking address to report: %v", err)
	}

	return addressID, nil
}

func (db *PortalRepository) SaveReportColumnByIdReport(idReport int, addressID int, infoReport map[string]string, index int) (int, error) {
	count := 0
	for name, value := range infoReport {
		query := `INSERT INTO report_column (report_id, id_address, name, value,index_column)
				  VALUES ($1, $2, $3 , $4 , $5)`
		result, err := db.Exec(query,  idReport, addressID,name, value, index)
		if err != nil {
			log.Printf("error saving report column: %v", err)
			return count, fmt.Errorf("error saving report column: %v", err)
		}
		rowsAffected, _ := result.RowsAffected()
		count += int(rowsAffected)
	}
	return count, nil
}

func (db *PortalRepository) SaveReportByIdUser(idUser int, nameReport string, instance string) (int, error) {
	var reportID int
	queryCheck := `SELECT id FROM report WHERE instance_hash = $1 AND name = $2`
	err := db.QueryRow(queryCheck, instance, nameReport).Scan(&reportID)
	if err == nil {
		log.Printf("report with instance_hash %s and name %s already exists, returning ID: %d", instance, nameReport, reportID)
		return reportID, nil
	}
	if err != sql.ErrNoRows {
		log.Printf("error checking existing report: %v", err)
		return 0, fmt.Errorf("error checking existing report: %v", err)
	}

	queryInsert := `INSERT INTO report (name, author, instance_hash)
				   VALUES ($1, $2, $3)
				   RETURNING id`
	err = db.QueryRow(queryInsert, nameReport, idUser, instance).Scan(&reportID)
	if err != nil {
		log.Printf("error saving report: %v", err)
		return 0, fmt.Errorf("error saving report: %v", err)
	}
	return reportID, nil
}

func (db PortalRepository) GetAddressInfoByUserId(userID int) ([]dto.AddressReport, error) {
	// Execute query
	query := `
		WITH AddressAttributes AS (
			SELECT
				ra.address_id,
				jsonb_object_agg(rc.name, rc.value ORDER BY rc.name) AS atributos
			FROM report_column rc
			JOIN report r ON r.id = rc.report_id
			JOIN report_address ra ON ra.report_id = r.id
			WHERE r.author = $1
			GROUP BY ra.address_id
		),
		AddressReport AS (
			WITH DedupedReports AS (
				SELECT DISTINCT
					ra.address_id,
					r.id AS report_id,
					r.name AS report_name
				FROM report r
				JOIN report_address ra ON ra.report_id = r.id
				WHERE r.author = $1
			)
			SELECT
				address_id,
				jsonb_agg(
					jsonb_build_object(
						'report_id', report_id,
						'report_name', report_name
					) ORDER BY report_id
				) AS report_details
			FROM DedupedReports
			GROUP BY address_id
			ORDER BY address_id
		)
		SELECT
			a.address AS address,
			a.normalized_address,
			a.latitude,
			a.longitude,
			jsonb_agg(
				jsonb_build_object(
					'report_id', ra.report_id,
					'report_name', r.name,
					'atributos', COALESCE(ra_attrs.atributos, '{}')
				) ORDER BY ra.report_id
			) AS atributos_relacionados,
			ar.report_details AS reportes
		FROM address a
		JOIN report_address ra ON ra.address_id = a.id
		JOIN report r ON r.id = ra.report_id
		LEFT JOIN AddressAttributes ra_attrs ON ra.address_id = ra_attrs.address_id
		LEFT JOIN AddressReport ar ON ra.address_id = ar.address_id
		WHERE r.author = $1
		GROUP BY a.address, a.normalized_address, a.latitude, a.longitude, ar.report_details
		ORDER BY a.address`

	rows, err := db.Query(query, userID)
	if err != nil {
		//log.Printf("Error querying reports: %v", err)
		//http.Error(w, "Internal server error", http.StatusInternalServerError)
		return nil,err
	}
	defer rows.Close()

	var reports []dto.AddressReport
	for rows.Next() {
		var (
			rpt       dto.AddressReport
			attrsJSON []byte
			repsJSON  []byte
		)
		err := rows.Scan(
			&rpt.Address,
			&rpt.NormalizedAddress,
			&rpt.Latitude,
			&rpt.Longitude,
			&attrsJSON,
			&repsJSON,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			//http.Error(w, "Internal server error", http.StatusInternalServerError)
			return  nil,err
		}
		// Unmarshal JSON fields
		if err := json.Unmarshal(attrsJSON, &rpt.AtributosRelacionados); err != nil {
			log.Printf("Error unmarshaling atributos_relacionados: %v", err)
			//http.Error(w, "Internal server error", http.StatusInternalServerError)
			return  nil,err
		}
		if err := json.Unmarshal(repsJSON, &rpt.Reportes); err != nil {
			log.Printf("Error unmarshaling reportes: %v", err)
			//http.Error(w, "Internal server error", http.StatusInternalServerError)
			return  nil,err
		}
		reports = append(reports, rpt)
	}
	return reports, nil
}
