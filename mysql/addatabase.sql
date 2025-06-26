-- MySQL dump 10.13  Distrib 8.0.34, for macos13 (arm64)
--
-- Host: 223.193.36.169    Database: addatabase
-- ------------------------------------------------------
-- Server version	8.0.11-TiDB-v7.5.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `aditems`
--

DROP TABLE IF EXISTS `aditems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `aditems` (
  `item_name` varchar(255) DEFAULT NULL,
  `redirect_url` varchar(255) DEFAULT NULL,
  `text` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `aditems`
--

LOCK TABLES `aditems` WRITE;
/*!40000 ALTER TABLE `aditems` DISABLE KEYS */;
INSERT INTO `aditems` VALUES ('lens','/product/66VCHSJNUP','Lens for sale. 50% off.'),('camera','/product/2ZYFJ3GM2N','Camera for sale. 20% off.'),('recordPlayer','/product/0PUK6V6EV0','Record player for sale. 30% off.'),('bike','/product/9SIQT8TOJO','City bike for sale. 10% off.'),('baristaKit','/product/1YMWWN1N4O','Barista Kit for sale. Buy one, get second kit for free'),('airPlant','/product/6E92ZMYYFZ','Air Plant for sale. Buy two, get third one for free'),('terrarium','/product/L9ECAV7KIM','Terrarium for sale. Buy one, get second one for free');
/*!40000 ALTER TABLE `aditems` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-07-17  9:40:25
