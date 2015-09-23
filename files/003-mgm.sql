# migrate estate tables from mgm to opensim database
# this is cross-database, and must be done manually by mgm

# rename users table to pending users for clarity
RENAME TABLE `users` TO `pendingusers`;

# remove transitory and unneccesary data from the hosts table
ALTER TABLE `hosts`
  DROP `port`,
  DROP `cmd_key`,
  DROP `status`;

# external address should be attached to a host record, not a region
ALTER TABLE  `hosts` ADD  `externalAddress` CHAR( 15 ) NOT NULL AFTER  `address` ;

# regions will reference hosts by id instead of address
ALTER TABLE  `regions` ADD  `host` INT NOT NULL DEFAULT  '0';

# populate new hosts column
UPDATE `regions` as t1 INNER JOIN `hosts` as t2 ON t1.slaveAddress = t2.address
  SET t1.host = IF(t1.slaveAddress is NULL, '0', t2.id);

# remove transitory data from the regions table
ALTER TABLE `regions`
  DROP `externalAddress`,
  DROP `slaveAddress`,
  DROP `isRunning`,
  DROP `status`;

DROP TABLE `migrations`;

INSERT IGNORE INTO `mgmDb` (`version`, `description`) VALUES (3, '003-mgm.sql');
