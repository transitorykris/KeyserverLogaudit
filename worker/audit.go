package main

import (
	"fmt"
)

// Audit performs and iteration of hash checking
func (w *Worker) Audit() error {
	// get a buffer to the actual log
	keyLog, err := getKeyserverLog(w.logURL)
	if err != nil {
		return err
	}
	previous := &Record{}
	last, err := w.lastRecord()
	if err != nil {
		return fmt.Errorf("failed to get last record from DB %v", err)
	}

	// save is used to indicate whether or not we're ready to start saving new records
	var save bool
	if last.Claimed == "" {
		w.log.Infoln("Our last record is the root record, start saving immediately")
		save = true
	} else {
		w.log.Infoln("Last record found:", last.Claimed)
	}

	// failed is used to flag that we've started to see bad records
	var failed bool

	// get the previous hashes from the DB to compare to the new hashes
	prevHash, err := w.getPreviousHashes()
	if err != nil {
		return fmt.Errorf("failed to retrieve previous hashes %v", err)
	}

	// audit our log

	id := 0 // id tracks our position in the log and database rows
	for {
		id++ // note: we start counting at 1 intentionally
		record, err := nextRecord(keyLog)
		if err != nil {
			// We've hit the end of the log
			break
		}

		// Generate our own hash for this record
		record.GenerateHash(previous)
		w.log.Debugln(record.Claimed, record.Hash)

		// If there is a hash mismatch this is bad!
		if record.Claimed != record.Hash {
			w.log.Errorln("Hash of record does not match the claim\n%+v\n", record)
			failed = true
		}

		// If we still have previous hashes to check this record against, and
		// if this hash doesn't match the hash of the previous run, that's bad too!
		if len(prevHash) > id && prevHash[id] != record.Hash {
			w.log.Errorln("Hash of record does not match the hash of the record from the previouis run %+v %s", record, prevHash[id])
			failed = true
		}

		if failed {
			if err = w.saveBadRecord(record, previous); err != nil {
				return fmt.Errorf("failed to save record %v", err)
			}
		} else if save {
			if err = w.saveRecord(record, previous); err != nil {
				return fmt.Errorf("failed to save record %v", err)
			}
		}

		if record.Claimed == last.Claimed {
			// This is the last record from the feed that we have in our DB
			// Start saving future records in the DB. If this never turns true
			// then the keyserver log has an integrity problem!
			save = true
		}

		// Save this record because we'll need its hash to check the next record
		previous = record
	}

	return nil
}
