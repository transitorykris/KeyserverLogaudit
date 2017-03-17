package main

// saveRun inserts an entry in the run table with some metadata about this run
func (w *Worker) saveRun() error {
	finalHash, err := w.lastRecord()
	if err != nil {
		return err
	}

	var records int
	if err = w.db.Get(&records, "SELECT COUNT(id) FROM `record`"); err != nil {
		return err
	}

	var badRecords int64
	if err = w.db.Get(&badRecords, "SELECT COUNT(id) FROM `bad_record`"); err != nil {
		return err
	}

	_, err = w.db.Exec("INSERT INTO `run` (`final_hash`, `record_count`, `bad_record`) VALUES (?, ?, ?)",
		finalHash.Claimed, records, badRecords,
	)

	return err
}
