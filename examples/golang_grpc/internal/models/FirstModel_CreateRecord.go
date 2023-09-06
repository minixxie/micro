package models

import ()

func (model *FirstModelImpl) CreateRecord(name string) (uint32, error) {

	// nowMS := time.Now().UnixNano() / 1000000
	var id uint32
	{
		stmt, err := model.MainDB.Prepare(`
			INSERT INTO test (random, name)
			VALUES (RAND(), ?)
		`)
		if err != nil {
			return 0, err
		}
		defer stmt.Close()

		insertResult, err := stmt.Exec(name)
		if err != nil {
			return 0, err
		}

		lastInsertId, err := insertResult.LastInsertId()
		if err != nil {
			return 0, err
		}
		id = uint32(lastInsertId)
	}
	return id, nil
}
