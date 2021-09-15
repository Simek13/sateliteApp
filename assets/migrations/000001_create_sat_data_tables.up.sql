CREATE TABLE IF NOT EXISTS `satellites` ( 
    `id` int NOT NULL AUTO_INCREMENT, 
    `name` varchar(32) NOT NULL UNIQUE,
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `measurements` ( 
    `id` int NOT NULL AUTO_INCREMENT,
    `filename` varchar(32), 
    `idSat` int, 
    `timestamp` varchar(32), 
    `ionoIndex` float, 
    `ndviIndex` float, 
    `radiationIndex` float, 
    `specificMeasurement` varchar(32),
    PRIMARY KEY (`id`),
    FOREIGN KEY (`idSat`) REFERENCES `satellites`(`id`) ON DELETE NO ACTION ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `computations` ( 
    `id` int NOT NULL AUTO_INCREMENT,
    `idSat` int, 
    `duration` varchar(32), 
    `maxIono` float, 
    `minIono` float, 
    `avgIono` float, 
    `maxNdvi` float, 
    `minNdvi` float, 
    `avgNdvi` float, 
    `maxRad` float, 
    `minRad` float, 
    `avgRad` float, 
    `maxSpec` float, 
    `minSpec` float, 
    `avgSpec` float,
    PRIMARY KEY (`id`),
    FOREIGN KEY (`idSat`) REFERENCES `satellites`(`id`) ON DELETE NO ACTION ON UPDATE NO ACTION
);

