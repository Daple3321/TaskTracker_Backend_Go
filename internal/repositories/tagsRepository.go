package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Daple3321/TaskTracker/internal/entity"
)

var ErrTagAlreadyExists = errors.New("tag already exists")

type TagsRepository struct {
	db *sql.DB
}

func NewTagsRepository(db *sql.DB) *TagsRepository {

	tg := TagsRepository{
		db: db,
	}

	return &tg
}

func (t *TagsRepository) GetTagsForUser(ctx context.Context, userId int) ([]entity.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	result := []entity.Tag{}

	rows, err := t.db.QueryContext(ctx, "SELECT * FROM tags WHERE user_id = ?", userId)
	if err != nil {
		slog.Error("Error on query", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fetchedTag entity.Tag
		err := rows.Scan(&fetchedTag.Id, &fetchedTag.UserId, &fetchedTag.Name)
		if err != nil {
			return nil, err
		}

		result = append(result, fetchedTag)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TagsRepository) GetTagsForTask(ctx context.Context, userId int, taskId int) ([]entity.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	result := []entity.Tag{}

	rows, err := t.db.QueryContext(ctx, `
		SELECT tg.id, tg.user_id, tg.name
		FROM task_tags tt
		INNER JOIN tags tg ON tg.id = tt.tag_id
		INNER JOIN tasks ts ON ts.id = tt.task_id
		WHERE tt.task_id = ? AND ts.user_id = ? AND tg.user_id = ts.user_id
	`, taskId, userId)
	if err != nil {
		slog.Error("error on query", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fetchedTag entity.Tag
		err := rows.Scan(&fetchedTag.Id, &fetchedTag.UserId, &fetchedTag.Name)
		if err != nil {
			return nil, err
		}

		result = append(result, fetchedTag)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *TagsRepository) SetTagForTask(ctx context.Context, userId int, taskId int, tagId int) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := t.GetTagById(ctx, userId, tagId)
	if err != nil {
		return err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO task_tags (task_id, tag_id) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, taskId, tagId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *TagsRepository) FindTagByName(ctx context.Context, userId int, name string) (*entity.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	row := t.db.QueryRowContext(ctx,
		"SELECT id, user_id, name FROM tags WHERE user_id = ? AND name = ?",
		userId,
		name,
	)
	var tag entity.Tag
	if err := row.Scan(&tag.Id, &tag.UserId, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (t *TagsRepository) GetTagById(ctx context.Context, userId int, tagId int) (*entity.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	fetchedTag := entity.Tag{}

	row := t.db.QueryRowContext(ctx,
		"SELECT id, user_id, name FROM tags WHERE user_id = ? AND id = ?",
		userId,
		tagId,
	)

	if err := row.Scan(&fetchedTag.Id, &fetchedTag.UserId, &fetchedTag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &fetchedTag, nil
}

func (t *TagsRepository) CreateTag(ctx context.Context, userId int, name string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO tags (user_id, name) VALUES(?,?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, userId, name)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return int(lastId), nil
}

func (t *TagsRepository) DeleteTag(ctx context.Context, userId int, tagId int) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := t.GetTagById(ctx, userId, tagId)
	if err != nil {
		return err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM tags WHERE user_id = ? AND id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userId, tagId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (t *TagsRepository) DeleteTagFromTask(ctx context.Context, userId int, tagId int, taskId int) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := t.GetTagById(ctx, userId, tagId)
	if err != nil {
		return err
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, taskId, tagId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
