-- Gochan MySQL startup/update script
-- DO NOT DELETE

CREATE TABLE IF NOT EXISTS `DBPREFIXannouncements` (
	`id` SERIAL,
	`subject` VARCHAR(45) NOT NULL DEFAULT '',
	`message` TEXT NOT NULL CHECK (message <> ''),
	`poster` VARCHAR(45) NOT NULL CHECK (poster <> ''),
	`timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `DBPREFIXappeals` (
	`id` SERIAL,
	`ban` INT(11) UNSIGNED NOT NULL CHECK (ban <> 0),
	`message` TEXT NOT NULL CHECK (message <> ''),
	`timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	`denied` BOOLEAN DEFAULT false,
	`staff_response` TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS DBPREFIXbanlist (
	`id` SERIAL,
	`allow_read` BOOLEAN DEFAULT TRUE,
	`ip` VARCHAR(45) NOT NULL DEFAULT '',
	`name` VARCHAR(255) NOT NULL DEFAULT '',
	`name_is_regex` BOOLEAN DEFAULT FALSE,
	`filename` VARCHAR(255) NOT NULL DEFAULT '',
	`file_checksum` VARCHAR(255) NOT NULL DEFAULT '',
	`boards` VARCHAR(255) NOT NULL DEFAULT '*',
	`staff` VARCHAR(50) NOT NULL DEFAULT '',
	`timestamp` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	`expires` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`permaban` BOOLEAN NOT NULL DEFAULT TRUE,
	`reason` VARCHAR(255) NOT NULL DEFAULT '',
	`type` SMALLINT NOT NULL DEFAULT 3,
	`staff_note` VARCHAR(255) NOT NULL DEFAULT '',
	`appeal_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`can_appeal` BOOLEAN NOT NULL DEFAULT true,
	PRIMARY KEY (id)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXbanlist`
	CHANGE IF EXISTS `banned_by` `staff` VARCHAR(50) NOT NULL DEFAULT '',
	CHANGE IF EXISTS `id` `id` SERIAL,
	CHANGE IF EXISTS `expires` `expires` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CHANGE IF EXISTS `boards` `boards` VARCHAR(255) NOT NULL DEFAULT '*',
	ADD COLUMN IF NOT EXISTS `type` TINYINT UNSIGNED NOT NULL DEFAULT 3,
	ADD COLUMN IF NOT EXISTS `name_is_regex` BOOLEAN DEFAULT FALSE,
	ADD COLUMN IF NOT EXISTS `filename` VARCHAR(255) NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS `file_checksum` VARCHAR(255) NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS `permaban` BOOLEAN DEFAULT FALSE,
	ADD COLUMN IF NOT EXISTS `can_appeal` BOOLEAN DEFAULT TRUE,
	DROP COLUMN IF EXISTS `message`;

DROP TABLE IF EXISTS `DBPREFIXbannedhashes`;

CREATE TABLE IF NOT EXISTS `DBPREFIXboards` (
	`id` SERIAL,
	`list_order` TINYINT UNSIGNED NOT NULL DEFAULT 0,
	`dir` VARCHAR(45) NOT NULL CHECK (dir <> ''),
	`type` TINYINT UNSIGNED NOT NULL DEFAULT 0,
	`upload_type` TINYINT UNSIGNED NOT NULL DEFAULT 0,
	`title` VARCHAR(45) NOT NULL CHECK (title <> ''),
	`subtitle` VARCHAR(64) NOT NULL DEFAULT '',
	`description` VARCHAR(64) NOT NULL DEFAULT '',
	`section` INT NOT NULL DEFAULT 1,
	`max_file_size` INT UNSIGNED NOT NULL DEFAULT 4718592,
	`max_pages` TINYINT UNSIGNED NOT NULL DEFAULT 11,
	`default_style` VARCHAR(45) NOT NULL DEFAULT '',
	`locked` BOOLEAN NOT NULL DEFAULT FALSE,
	`created_on` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`anonymous` VARCHAR(45) NOT NULL DEFAULT 'Anonymous',
	`forced_anon` BOOLEAN NOT NULL DEFAULT FALSE,
	`max_age` INT(20) UNSIGNED NOT NULL DEFAULT 0,
	`autosage_after` INT(5) UNSIGNED NOT NULL DEFAULT 200,
	`no_images_after` INT(5) UNSIGNED NOT NULL DEFAULT 0,
	`max_message_length` INT(10) UNSIGNED NOT NULL DEFAULT 8192,
	`embeds_allowed` BOOLEAN NOT NULL DEFAULT TRUE,
	`redirect_to_thread` BOOLEAN NOT NULL DEFAULT TRUE,
	`require_file` BOOLEAN NOT NULL DEFAULT FALSE,
	`enable_catalog` BOOLEAN NOT NULL DEFAULT TRUE,
	PRIMARY KEY (`id`),
	UNIQUE (`dir`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXboards`
	CHANGE COLUMN IF EXISTS `order` `list_order` INT UNSIGNED NOT NULL DEFAULT 0,
	CHANGE COLUMN IF EXISTS `max_image_size` `max_file_size` INT UNSIGNED NOT NULL DEFAULT 4718592,
	CHANGE COLUMN IF EXISTS `default_style` `default_style` VARCHAR(45) NOT NULL DEFAULT '',
	DROP COLUMN IF EXISTS `locale`;

CREATE TABLE IF NOT EXISTS `DBPREFIXembeds` (
	`id` SERIAL,
	`filetype` VARCHAR(3) NOT NULL,
	`name` VARCHAR(45) NOT NULL,
	`video_url` VARCHAR(255) NOT NULL,
	`width` SMALLINT UNSIGNED NOT NULL,
	`height` SMALLINT UNSIGNED NOT NULL,
	`embed_code` TEXT NOT NULL,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `DBPREFIXDBPREFIXfrontpage`;

CREATE TABLE IF NOT EXISTS `DBPREFIXinfo` (
	`name` VARCHAR(45) NOT NULL,
	`value` TEXT NOT NULL,
	PRIMARY KEY (`name`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `DBPREFIXlinks` (
	`id` SERIAL,
	`title` VARCHAR(45) NOT NULL,
	`url` VARCHAR(255) NOT NULL,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS DBPREFIXloginattempts;
DROP TABLE IF EXISTS DBPREFIXpluginsettings;

CREATE TABLE IF NOT EXISTS `DBPREFIXposts` (
	`id` SERIAL,
	`boardid` INT NOT NULL,
	`parentid` INT(10) UNSIGNED NOT NULL DEFAULT '0',
	`name` VARCHAR(50) NOT NULL,
	`tripcode` VARCHAR(10) NOT NULL,
	`email` VARCHAR(50) NOT NULL,
	`subject` VARCHAR(100) NOT NULL,
	`message` TEXT NOT NULL,
	`message_raw` TEXT NOT NULL,
	`password` VARCHAR(45) NOT NULL,
	`filename` VARCHAR(45) NOT NULL DEFAULT '',
	`filename_original` VARCHAR(255) NOT NULL DEFAULT '',
	`file_checksum` VARCHAR(45) NOT NULL DEFAULT '',
	`filesize` INT(20) UNSIGNED NOT NULL DEFAULT 0,
	`image_w` SMALLINT(5) UNSIGNED NOT NULL DEFAULT 0,
	`image_h` SMALLINT(5) UNSIGNED NOT NULL DEFAULT 0,
	`thumb_w` SMALLINT(5) UNSIGNED NOT NULL DEFAULT 0,
	`thumb_h` SMALLINT(5) UNSIGNED NOT NULL DEFAULT 0,
	`ip` VARCHAR(45) NOT NULL DEFAULT '',
	`tag` VARCHAR(5) NOT NULL DEFAULT '',
	`timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`autosage` BOOLEAN NOT NULL DEFAULT FALSE,
	`deleted_timestamp` TIMESTAMP,
	`bumped` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`stickied` BOOLEAN NOT NULL DEFAULT FALSE,
	`locked` BOOLEAN NOT NULL DEFAULT FALSE,
	`reviewed` BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY  (`boardid`,`id`),
	KEY `parentid` (`parentid`),
	KEY `bumped` (`bumped`),
	KEY `file_checksum` (`file_checksum`),
	KEY `stickied` (`stickied`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXposts`
	DROP COLUMN IF EXISTS `sillytag`,
	DROP COLUMN IF EXISTS `poster_authority`;

CREATE TABLE IF NOT EXISTS `DBPREFIXreports` (
	`id` SERIAL,
	`board` VARCHAR(45) NOT NULL,
	`postid` INT(10) UNSIGNED NOT NULL,
	`timestamp` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`ip` VARCHAR(45) NOT NULL,
	`reason` VARCHAR(255) NOT NULL,
	`cleared` BOOLEAN NOT NULL DEFAULT FALSE,
	`istemp` BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `DBPREFIXsections` (
	`id` SERIAL,
	`list_order` INT UNSIGNED NOT NULL DEFAULT 0,
	`hidden` BOOLEAN NOT NULL DEFAULT FALSE,
	`name` VARCHAR(45) NOT NULL,
	`abbreviation` VARCHAR(10) NOT NULL,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXsections`
	CHANGE COLUMN IF EXISTS `order` `list_order` INT UNSIGNED NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS `DBPREFIXsessions` (
	`id` SERIAL,
	`name` CHAR(16) NOT NULL,
	`sessiondata` VARCHAR(45) NOT NULL,
	`expires` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXsessions`
	CHANGE IF EXISTS `key` `name` CHAR(16) NOT NULL,
	CHANGE IF EXISTS `data` `sessiondata` VARCHAR(45) NOT NULL;

CREATE TABLE IF NOT EXISTS `DBPREFIXstaff` (
	`id` SERIAL,
	`username` VARCHAR(45) NOT NULL,
	`password_checksum` VARCHAR(120) NOT NULL,
	`rank` TINYINT(1) UNSIGNED NOT NULL DEFAULT 2,
	`boards` VARCHAR(128) NOT NULL DEFAULT '*',
	`added_on` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`last_active` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`),
	UNIQUE (`username`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `DBPREFIXstaff`
	CHANGE IF EXISTS `boards` `boards` VARCHAR(128) NOT NULL DEFAULT '*',
	DROP COLUMN IF EXISTS `salt`;

CREATE TABLE IF NOT EXISTS `DBPREFIXwordfilters` (
	`id` SERIAL,
	`search` VARCHAR(75) NOT NULL CHECK (search <> ''),
	`change_to` VARCHAR(75) NOT NULL DEFAULT '',
	`boards` VARCHAR(128) NOT NULL DEFAULT '*',
	`regex` BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY (`id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
