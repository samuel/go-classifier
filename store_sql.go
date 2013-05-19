package classifier

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	categoriesTable               = "categories"
	tokensTable                   = "tokens"
	categoriesQuery               = `SELECT "id", "name", "document_count" FROM ` + categoriesTable
	insertCategoryQuery           = `INSERT INTO ` + categoriesTable + ` ("name", "document_count") VALUES (?, 0)`
	updateDocCountQuery           = `UPDATE ` + categoriesTable + ` SET document_count = document_count + ? WHERE "name" = ?`
	updateOrInsertTokenCountQuery = `INSERT OR REPLACE INTO ` + tokensTable + ` ("category_id", "token", "count") VALUES (?, ?, ? + COALESCE((SELECT "count" FROM ` + tokensTable + ` WHERE "category_id" = ? AND "token" = ?), 0))`
	updateTokenCountQuery         = `UPDATE ` + tokensTable + ` SET "count" = "count" + ? WHERE "category_id" = ? AND "token" = ? AND "count" > 0`
	tokensQuery                   = `SELECT "id", "category_id", "token", "count" FROM ` + tokensTable + ` WHERE category_id IN (%s) AND token IN (%s)`
)

type SQLStore struct {
	db                  *sql.DB
	categoriesQuery     *sql.Stmt
	insertCategoryQuery *sql.Stmt
	tokensQuery         *sql.Stmt
}

func NewSQLStore(db *sql.DB) (*SQLStore, error) {
	s := &SQLStore{
		db: db,
	}
	var err error
	s.categoriesQuery, err = db.Prepare(categoriesQuery)
	if err != nil {
		return nil, err
	}
	s.insertCategoryQuery, err = db.Prepare(insertCategoryQuery)
	return s, err
}

func (s *SQLStore) Categories() (map[string]int64, error) {
	rows, err := s.categoriesQuery.Query()
	if err != nil {
		return nil, err
	}
	categories := make(map[string]int64, 0)
	for rows.Next() {
		var id int64
		var name string
		var documentCount int64
		if err := rows.Scan(&id, &name, &documentCount); err != nil {
			return nil, err
		}
		categories[name] = documentCount
	}
	return categories, rows.Err()
}

func (s *SQLStore) AddCategory(name string) error {
	res, err := s.insertCategoryQuery.Exec(name)
	if err != nil {
		return err
	}
	_ = res
	return err
}

func (s *SQLStore) AddDocument(category string, tokens []string) error {
	return s.updateDocument(category, tokens, 1)
}

func (s *SQLStore) RemoveDocument(category string, tokens []string) error {
	return s.updateDocument(category, tokens, -1)
}

func (s *SQLStore) updateDocument(category string, tokens []string, delta int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec(updateDocCountQuery, delta, category)
	if err != nil {
		tx.Rollback()
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return ErrCategoryDoesNotExist(category)
	}
	row := tx.QueryRow("SELECT id FROM categories WHERE name = ?", category)
	var categoryId int64
	if err := row.Scan(&categoryId); err != nil {
		tx.Rollback()
		return err
	}
	for _, t := range tokens {
		var res sql.Result
		if delta > 0 {
			res, err = tx.Exec(updateOrInsertTokenCountQuery, categoryId, t, delta, categoryId, t)
		} else {
			res, err = tx.Exec(updateTokenCountQuery, delta, categoryId, t)
		}
		if err != nil {
			tx.Rollback()
			return err
		} else if n, err := res.RowsAffected(); err != nil {
			tx.Rollback()
			return err
		} else if n != 1 {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return errors.New("classifier: failed to update token")
		}
	}
	tx.Commit()
	return nil
}

// TODO: make this more robust and secure (database specific?)
func (s *SQLStore) escape(st string) string {
	st = strings.Replace(st, "'", "''", -1)
	return "'" + st + "'"
}

func (s *SQLStore) TokenCounts(categories []string, tokens []string) (map[string]map[string]int64, error) {
	rows, err := s.categoriesQuery.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categoryMap := make(map[string]int64, len(categories))
	revCategoryMap := make(map[int64]string, len(categories))
	for rows.Next() {
		var id int64
		var name string
		var documentCount int64
		if err := rows.Scan(&id, &name, &documentCount); err != nil {
			return nil, err
		}
		categoryMap[name] = id
		revCategoryMap[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	categoryIds := make([]string, len(categories))
	if categories == nil {
		categories = make([]string, 0, len(categoryMap))
		for name, id := range categoryMap {
			categories = append(categories, name)
			categoryIds = append(categoryIds, strconv.FormatInt(id, 10))
		}
	} else {
		for i, c := range categories {
			if id, ok := categoryMap[c]; !ok {
				return nil, ErrCategoryDoesNotExist(c)
			} else {
				categoryIds[i] = strconv.FormatInt(id, 10)
			}
		}
	}
	escapedTokens := make([]string, len(tokens))
	for i, t := range tokens {
		escapedTokens[i] = s.escape(t)
	}
	query := fmt.Sprintf(tokensQuery, strings.Join(categoryIds, ","), strings.Join(escapedTokens, ","))
	rows, err = s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]map[string]int64, len(categories))
	for _, c := range categories {
		res[c] = make(map[string]int64, 0)
	}
	for rows.Next() {
		var id int64
		var categoryId int64
		var token string
		var count int64
		if err := rows.Scan(&id, &categoryId, &token, &count); err != nil {
			return nil, err
		}
		if c := revCategoryMap[categoryId]; c != "" {
			res[c][token] = count
		}
	}
	return res, rows.Err()
}
