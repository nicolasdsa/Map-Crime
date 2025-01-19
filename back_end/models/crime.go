package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"crimemap.com/map/database"
	"crimemap.com/map/dto"
	"github.com/jackc/pgx/v5"
)

type Crime struct {
	ID             int            `json:"id"`
	CrimeDate      time.Time      `json:"crime_date"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	SubcategoriaId      int    		`json:"subcategory_id"`
}

func GetCrimes(ctx context.Context, conn *pgx.Conn, filter dto.CrimeFilter) ([]Crime, error) {
	conn = database.Connect()
	if conn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	defer conn.Close(ctx)

	query := `
        SELECT 
            id,
            ST_X(geom) AS longitude,
            ST_Y(geom) AS latitude,
            crime_date,
						subcategory_id
        FROM crimes
        WHERE ST_Within(
            geom,
            ST_MakeEnvelope($1, $2, $3, $4, 4326)
        )
        AND crime_date >= $5
        AND crime_date <= $6
    `

	args := []interface{}{
		filter.West, filter.South, filter.East, filter.North,
		filter.StartDate, filter.EndDate,
	}

	// Adiciona lógica para exclusão de IDs
	if len(filter.ExcludedIDs) > 0 {
		query += " AND id != ALL($7)"
		args = append(args, filter.ExcludedIDs)
	} else {
		// Passa um array vazio explicitamente tipado
		query += " AND id != ALL('{}'::int[])"
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var crimes []Crime
	for rows.Next() {
		var crime Crime
		if err := rows.Scan(&crime.ID, &crime.Longitude, &crime.Latitude, &crime.CrimeDate, &crime.SubcategoriaId); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		crimes = append(crimes, crime)
	}

	if rows.Err() != nil {
		log.Printf("Rows error: %v", rows.Err())
		return nil, rows.Err()
	}

	fmt.Println(crimes)

	return crimes, nil
}

