
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE `record` (
    `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
    `seen` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `claimed_hash` VARCHAR(64) DEFAULT "",
    `previous_hash` VARCHAR(64) DEFAULT "",
    `actual_hash` VARCHAR(64) DEFAULT "",
    `text` TEXT
);

INSERT INTO `record` (`text`) VALUES ("");

CREATE TABLE `bad_record` (
    `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
    `seen` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `claimed_hash` VARCHAR(64) DEFAULT "",
    `previous_hash` VARCHAR(64) DEFAULT "",
    `actual_hash` VARCHAR(64) DEFAULT "",
    `text` TEXT
);

CREATE TABLE `run` (
    `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
    `final_hash` VARCHAR(64) DEFAULT "",
    `record_count` INTEGER DEFAULT 0,
    `bad_record` INTEGER DEFAULT 0,
    `date` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO `run` () VALUES ();

CREATE TABLE `email` (
    `id` INTEGER PRIMARY KEY AUTO_INCREMENT,
    `address` VARCHAR(255) UNIQUE,
    `code` VARCHAR(64),
    `confirmed` BOOLEAN DEFAULT false
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE `email`;

DROP TABLE `run`;

DROP TABLE `bad_record`;

DROP TABLE `record`;