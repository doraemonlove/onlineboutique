-- MySQL dump 10.13  Distrib 8.0.34, for macos13 (arm64)
--
-- Host: 223.193.36.169    Database: productdb
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
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `picture` varchar(255) DEFAULT NULL,
  `priceUsd` json DEFAULT NULL,
  `categories` json DEFAULT NULL,
  PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `products`
--

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES ('0PUK6V6EV0','Vintage Record Player','It still works.','/static/img/products/record-player.jpg','{\"currencyCode\": \"USD\", \"nanos\": 500000000, \"units\": 65}','[\"music\", \"vintage\"]'),('1YMWWN1N4O','Home Barista Kit','Always wanted to brew coffee with Chemex and Aeropress at home?','/static/img/products/barista-kit.jpg','{\"currencyCode\": \"USD\", \"nanos\": 0, \"units\": 124}','[\"cookware\"]'),('2ZYFJ3GM2N','Film Camera','This camera looks like it\'s a film camera, but it\'s actually digital.','/static/img/products/film-camera.jpg','{\"currencyCode\": \"USD\", \"nanos\": 0, \"units\": 2245}','[\"photography\", \"vintage\"]'),('66VCHSJNUP','Vintage Camera Lens','You won\'t have a camera to use it and it probably doesn\'t work anyway.','/static/img/products/camera-lens.jpg','{\"currencyCode\": \"USD\", \"nanos\": 490000000, \"units\": 12}','[\"photography\", \"vintage\"]'),('6E92ZMYYFZ','Air Plant','Have you ever wondered whether air plants need water? Buy one and figure out.','/static/img/products/air-plant.jpg','{\"currencyCode\": \"USD\", \"nanos\": 300000000, \"units\": 12}','[\"gardening\"]'),('9SIQT8TOJO','City Bike','This single gear bike probably cannot climb the hills of San Francisco.','/static/img/products/city-bike.jpg','{\"currencyCode\": \"USD\", \"nanos\": 500000000, \"units\": 789}','[\"cycling\"]'),('L9ECAV7KIM','Terrarium','This terrarium will looks great in your white painted living room.','/static/img/products/terrarium.jpg','{\"currencyCode\": \"USD\", \"nanos\": 450000000, \"units\": 36}','[\"gardening\"]'),('LS4PSXUNUM','Metal Camping Mug','You probably don\'t go camping that often but this is better than plastic cups.','/static/img/products/camp-mug.jpg','{\"currencyCode\": \"USD\", \"nanos\": 330000000, \"units\": 24}','[\"cookware\"]'),('OLJCESPC7Z','Vintage Typewriter','This typewriter looks good in your living room.','/static/img/products/typewriter.jpg','{\"currencyCode\": \"USD\", \"nanos\": 990000000, \"units\": 67}','[\"vintage\"]');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-07-17  9:40:46
