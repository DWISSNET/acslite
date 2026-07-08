package services

import (
	"database/sql"
	"fmt"

	"github.com/DWISSNET/acsgo/models"
)

// ParameterService handles parameter template operations
type ParameterService struct {
	db *sql.DB
}

func NewParameterService(db *sql.DB) *ParameterService {
	return &ParameterService{db: db}
}

// Count returns the number of seeded parameters.
func (s *ParameterService) Count() (int, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM parameters").Scan(&n)
	return n, err
}

// Upsert inserts or updates a parameter template.
func (s *ParameterService) Upsert(p *models.Parameter) error {
	vendors, _ := p.SupportedVendors.Value()
	vmodels, _ := p.SupportedModels.Value()
	writable := 0
	if p.Writable {
		writable = 1
	}
	std := 0
	if p.IsStandardTR069 {
		std = 1
	}
	defVal := fmt.Sprintf("%v", p.DefaultValue)
	_, err := s.db.Exec(`
		INSERT INTO parameters (path, name, description, type, category, subcategory,
			writable, default_value, supported_vendors, supported_models, is_standard_tr069)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(path) DO UPDATE SET
			name               = excluded.name,
			description        = excluded.description,
			type               = excluded.type,
			category           = excluded.category,
			subcategory        = excluded.subcategory,
			writable           = excluded.writable,
			default_value      = excluded.default_value,
			supported_vendors  = excluded.supported_vendors,
			supported_models   = excluded.supported_models,
			is_standard_tr069  = excluded.is_standard_tr069
	`, p.Path, p.Name, p.Description, p.Type, p.Category, p.Subcategory,
		writable, defVal, vendors, vmodels, std)
	return err
}

// List returns all parameters, optionally filtered by category and/or vendor.
func (s *ParameterService) List(category, vendor string) ([]*models.Parameter, error) {
	query := `SELECT path, name, description, type, category, subcategory,
		writable, default_value, supported_vendors, supported_models, is_standard_tr069
		FROM parameters WHERE 1=1`
	var args []interface{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if vendor != "" {
		query += ` AND (supported_vendors LIKE ? OR supported_vendors LIKE '%"All"%')`
		args = append(args, `%"`+vendor+`"%`)
	}
	query += " ORDER BY path ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParameters(rows)
}

// GetByPath fetches a single parameter by its path.
func (s *ParameterService) GetByPath(path string) (*models.Parameter, error) {
	row := s.db.QueryRow(`SELECT path, name, description, type, category, subcategory,
		writable, default_value, supported_vendors, supported_models, is_standard_tr069
		FROM parameters WHERE path = ?`, path)

	params, err := scanParameters(row)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return nil, fmt.Errorf("parameter not found")
	}
	return params[0], nil
}

// Categories returns all unique category names.
func (s *ParameterService) Categories() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT category FROM parameters ORDER BY category ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// --------------------------------------------------------------------------
// scanner helpers
// --------------------------------------------------------------------------

type paramScanner interface {
	Scan(dest ...interface{}) error
}

func scanParameters(rows interface{}) ([]*models.Parameter, error) {
	var results []*models.Parameter
	scan := func(s paramScanner) (*models.Parameter, error) {
		var p models.Parameter
		var writable, std int
		var defVal string
		err := s.Scan(&p.Path, &p.Name, &p.Description, &p.Type, &p.Category, &p.Subcategory,
			&writable, &defVal, &p.SupportedVendors, &p.SupportedModels, &std)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		p.Writable = writable == 1
		p.IsStandardTR069 = std == 1
		p.DefaultValue = defVal
		return &p, nil
	}

	switch r := rows.(type) {
	case *sql.Rows:
		for r.Next() {
			p, err := scan(r)
			if err != nil {
				return nil, err
			}
			if p != nil {
				results = append(results, p)
			}
		}
	case *sql.Row:
		p, err := scan(r)
		if err != nil {
			return nil, err
		}
		if p != nil {
			results = append(results, p)
		}
	}
	return results, nil
}
