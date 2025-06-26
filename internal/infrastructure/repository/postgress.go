package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
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
		host = "54.156.84.208"
		log.Println("POSTGRES_HOST no definido, usando 'localhost'")
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
		log.Println("POSTGRES_PORT no definido, usando '5432'")
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "pontupin"
		log.Println("POSTGRES_USER no definido, usando 'postgres'")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "iddqd"
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
        WHERE s.token = $1 AND s.is_active = true 
    ` //AND s.expires_at > CURRENT_TIMESTAMP
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
	log.Println("Session logged successfully %s", active)
}

func (db *PortalRepository) SaveAddress(idReport int, address string, latitude float64, longitude float64, formatAddress string, geocoder string) (int, error) {
	var addressID int
	queryCheck := `SELECT id,address,normalized_address FROM address WHERE address = $1`
	addressDB := ""
	formatAddressDB := ""
	err := db.QueryRow(queryCheck, address).Scan(&addressID, &addressDB, &formatAddressDB)
	if err == nil {
		//Cuando hay un resultado de latitud y longitud, se actualiza la dirección
		if addressDB != address || formatAddressDB != formatAddress {
			updateQuery := `UPDATE address SET address = $1, normalized_address = $2, latitude = $3, longitude = $4, geocoder = $5 WHERE id = $6`
			_, err = db.Exec(updateQuery, address, formatAddress, latitude, longitude, geocoder, addressID)
		} else {
			_, err = db.SaveAddressInReport(idReport, addressID, latitude, longitude, formatAddress, geocoder)
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
		result, err := db.Exec(query, idReport, addressID, name, value, index)
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
		SELECT 
			a.address,
			a.normalized_address,
			a.latitude,
			a.longitude,
			(
				SELECT json_agg(attr_agg)
				FROM (
					SELECT json_build_object(
						'atributos', json_object_agg(rc.name, rc.value)
					) AS attr_agg
					FROM public.report_address ra2
					JOIN public.report_column rc ON rc.report_id = ra2.report_id
					JOIN public.report r2 ON r2.id = ra2.report_id
					WHERE ra2.address_id = a.id
					AND r2.author = $1
					GROUP BY rc.report_id
				) AS sub_attr
			) AS atributos_relacionados,
			(
				SELECT json_agg(DISTINCT jsonb_build_object(
					'report_id', r.id,
					'report_name', r.name
				))
				FROM public.report_address ra
				JOIN public.report r ON r.id = ra.report_id
				WHERE ra.address_id = a.id
				AND r.author = $1
			) AS reportes
		FROM public.address a
		JOIN public.report_address ra ON ra.address_id = a.id
		JOIN public.report r ON r.id = ra.report_id
		JOIN public.users u ON u.id = r.author
		WHERE r.author = $1
		AND a.latitude != 0
		AND a.longitude != 0
		GROUP BY a.id, a.address, a.normalized_address, a.latitude, a.longitude
		`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		if err := json.Unmarshal(attrsJSON, &rpt.AtributosRelacionados); err != nil {
			log.Printf("Error unmarshaling atributos_relacionados: %v", err)
			return nil, err
		}
		if err := json.Unmarshal(repsJSON, &rpt.Reportes); err != nil {
			log.Printf("Error unmarshaling reportes: %v", err)
			return nil, err
		}
		reports = append(reports, rpt)
	}
	return reports, nil
}

func (db PortalRepository) GetReportSummaryByUserId(userID int) ([]dto.ReportResume, error) {
	query := `
        SELECT 
			r.id,
            r."name",
            r.created_at,
			r.status,
            COUNT(*) AS direcciones
        FROM
            report r
            JOIN report_address ra ON r.id = ra.report_id
        WHERE
            r.author = $1
        GROUP BY 
			r.id,
            r."name",
            r.created_at
		order by  r.created_at desc
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []dto.ReportResume
	for rows.Next() {
		var summary dto.ReportResume
		err := rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.CreatedAt,
			&summary.Status,
			&summary.Direcciones,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (db PortalRepository) GetReportRowsByReportID(reportID int, page int, pageSize int) ([]dto.ReportRow, int, error) {
	// Consulta para contar el total de filas
	countQuery := `
        SELECT COUNT(DISTINCT rc.index_column)
        FROM report r
        LEFT JOIN report_column rc ON r.id = rc.report_id
        WHERE r.id = $1
    `
	var totalRows int
	err := db.QueryRow(countQuery, reportID).Scan(&totalRows)
	if err != nil {
		log.Printf("Error counting report rows: %v", err)
		return nil, 0, err

	}

	// Consulta para obtener las filas paginadas
	query := `
        SELECT 
            rc.index_column,
            json_object_agg(rc.name, rc.value) AS fila_transpuesta
        FROM report r
        LEFT JOIN report_column rc ON r.id = rc.report_id
        WHERE r.id = $1
        GROUP BY rc.index_column
        ORDER BY rc.index_column
        LIMIT $2 OFFSET $3
    `

	rows, err := db.Query(query, reportID, pageSize, page*pageSize)
	if err != nil {
		log.Printf("Error querying report rows: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var reportRows []dto.ReportRow
	for rows.Next() {
		var (
			row      dto.ReportRow
			filaJSON []byte
		)
		err := rows.Scan(
			&row.IndexColumn,
			&filaJSON,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, 0, err
		}

		if err := json.Unmarshal(filaJSON, &row.FilaTranspuesta); err != nil {
			log.Printf("Error unmarshaling fila_transpuesta: %v", err)
			return nil, 0, err
		}

		reportRows = append(reportRows, row)
	}

	return reportRows, totalRows, nil
}
func (db *PortalRepository) GetTotalReportsAndAddress(userID int) ([]dto.CategoryCount, error) {
	query := `
        SELECT 
            'address' AS category,
            COUNT(DISTINCT a.id) AS total
        FROM 
            address a
            JOIN report_address ra ON a.id = ra.address_id
            JOIN report r ON ra.report_id = r.id
        WHERE
            r.author = $1
        UNION
        SELECT 
            'report' AS category,
            COUNT(*) AS total
        FROM 
            report r 
        WHERE
            r.author = $1
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("Error querying reports and addresses: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []dto.CategoryCount
	for rows.Next() {
		var result dto.CategoryCount
		err := rows.Scan(
			&result.Category,
			&result.Total,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (db PortalRepository) GetAddressInfoByUserIdPeerPage(userID int, query string, limit, offset int) ([]dto.AddressReport, int, error) {
	var wg sync.WaitGroup
	var addresses []dto.AddressReport
	var total int
	var errCount, errQuery error

	wg.Add(2)

	go func() {
		defer wg.Done()
		countQuery := `
            SELECT COUNT(DISTINCT a.id)
            FROM address a
            INNER JOIN report_address ra ON ra.address_id = a.id
            INNER JOIN report r ON ra.report_id = r.id
            WHERE r.author = $1
            AND (
                $2 = ''
                OR a.address ILIKE '%' || $2 || '%'
                OR a.normalized_address ILIKE '%' || $2 || '%'
            )
        `
		errCount = db.QueryRow(countQuery, userID, query).Scan(&total)
		if errCount != nil {
			log.Printf("Error counting addresses: %v", errCount)
		}
	}()

	// Goroutine para obtener las direcciones
	go func() {
		defer wg.Done()
		querySQL := `
            SELECT 
                a.id,
                a.address,
                a.normalized_address,
                a.latitude,
                a.longitude,
                COALESCE(
                    (
                        SELECT json_agg(attrs)
                        FROM (
                            SELECT json_build_object(
                                'atributos', json_object_agg(rc.name, rc.value)
                            ) AS attrs
                            FROM report_column rc
                            INNER JOIN report_address ra2 ON ra2.report_id = rc.report_id
                            WHERE ra2.address_id = a.id
                            GROUP BY rc.report_id
                        ) sub
                    ),
                    '[]'
                ) AS atributos_relacionados,
                COALESCE(
                    (
                        SELECT json_agg(
                            json_build_object(
                                'report_id', r2.id,
                                'report_name', r2.name
                            )
                        )
                        FROM report r2
                        INNER JOIN report_address ra2 ON ra2.report_id = r2.id
                        WHERE ra2.address_id = a.id
                    ),
                    '[]'
                ) AS reportes
            FROM address a
            INNER JOIN report_address ra ON ra.address_id = a.id
            INNER JOIN report r ON ra.report_id = r.id
            WHERE r.author = $1
            AND (
                $2 = ''
                OR a.address ILIKE '%' || $2 || '%'
                OR a.normalized_address ILIKE '%' || $2 || '%'
            )
            GROUP BY a.id, a.address, a.normalized_address, a.latitude, a.longitude
            ORDER BY a.id
            LIMIT $3 OFFSET $4
        `

		rows, err := db.Query(querySQL, userID, query, limit, offset)
		if err != nil {
			log.Printf("Error querying addresses: %v", err)
			errQuery = err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var (
				addr          dto.AddressReport
				atributosJSON []byte
				reportesJSON  []byte
			)
			err := rows.Scan(
				&addr.ID,
				&addr.Address,
				&addr.NormalizedAddress,
				&addr.Latitude,
				&addr.Longitude,
				&atributosJSON,
				&reportesJSON,
			)
			if err != nil {
				log.Printf("Error scanning address: %v", err)
				errQuery = err
				return
			}

			if err := json.Unmarshal(atributosJSON, &addr.AtributosRelacionados); err != nil {
				log.Printf("Error unmarshaling atributos: %v", err)
				errQuery = err
				return
			}

			if err := json.Unmarshal(reportesJSON, &addr.Reportes); err != nil {
				log.Printf("Error unmarshaling reportes: %v", err)
				errQuery = err
				return
			}

			addresses = append(addresses, addr)
		}
	}()

	wg.Wait()

	if errCount != nil {
		return nil, 0, errCount
	}
	if errQuery != nil {
		return nil, 0, errQuery
	}

	return addresses, total, nil
}

func (db *PortalRepository) GetReportByReportUserID(userID, reportID int) (dto.ReportResume, error) {
	query := `
		SELECT 
			r.name,
			r.id,
			r.status,
			r.created_at
		FROM report r
		WHERE
			r.id = $1
			AND r.author = $2
		order by r.created_at
	`

	var report dto.ReportResume
	var id int
	var status int
	err := db.QueryRow(query, reportID, userID).Scan(
		&report.Name,
		&id,
		&status,
		&report.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Report not found for userID %d and reportID %d", userID, reportID)
			return dto.ReportResume{}, fmt.Errorf("report not found for userID %d and reportID %d", userID, reportID)
		}
		log.Printf("Error querying report: %v", err)
		return dto.ReportResume{}, fmt.Errorf("error querying report: %w", err)
	}

	// Convertir el ID a string
	report.ID = fmt.Sprintf("%d", id)
	report.Status = status

	return report, nil
}

func (db *PortalRepository) FindAddress(address string) (dto.WeMapsAddress, error) {
	var geo dto.WeMapsAddress
	var similarityNormalized, similarityRaw float64

	// Consulta única que compara con ambos campos y selecciona el mejor puntaje
	err := db.QueryRow(`
		SELECT normalized_address, latitude, longitude, 
		       similarity($1, normalized_address) AS sim_normalized,
		       similarity($1, address) AS sim_raw
		FROM address
		ORDER BY GREATEST(similarity($1, normalized_address), similarity($1, address)) DESC
		LIMIT 1
	`, address).Scan(&geo.FormattedAddress, &geo.Latitude, &geo.Longitude, &similarityNormalized, &similarityRaw)
	if err != nil {
		return dto.WeMapsAddress{}, fmt.Errorf("error finding address: %v", err)
	}

	// Verificar si al menos un puntaje de similitud supera el umbral
	if similarityNormalized > 0.8 || similarityRaw > 0.8 {
		return geo, nil
	}

	// Si ningún puntaje supera el umbral
	return dto.WeMapsAddress{}, fmt.Errorf("puntaje de dirección muy bajo: normalized_address (%.2f), address (%.2f)", similarityNormalized, similarityRaw)
}

func (db *PortalRepository) FindUserByID(userID int) (*User, error) {
	query := `
        SELECT u.id, u.email, u.alias, u.full_name, u.phone
        FROM users u 
        WHERE u.id = $1 
    `
	var user User
	err := db.QueryRow(query, userID).Scan(
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

func (db *PortalRepository) SetStatusReport(userID, reportID, status int) (dto.ReportResume, error) {
	var report dto.ReportResume
	query := `
        UPDATE public.report
        SET status = $1
        WHERE id = $2 AND author = $3
        RETURNING id, name, author, instance_hash, created_at, status
    `

	err := db.DB.QueryRow(query, status, reportID, userID).Scan(
		&report.ID,
		&report.Name,
		&report.CreatedAt,
		&report.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return report, fmt.Errorf("no report found with id %d for user %d", reportID, userID)
		}
		return report, fmt.Errorf("failed to update report status: %w", err)
	}

	return report, nil
}
