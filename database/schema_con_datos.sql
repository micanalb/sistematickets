-- MySQL dump 10.13  Distrib 9.4.0, for Win64 (x86_64)
--
-- Host: localhost    Database: sistematickets
-- ------------------------------------------------------
-- Server version	9.4.0

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
-- Table structure for table `entradas`
--

DROP TABLE IF EXISTS `entradas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `entradas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `codigo_qr` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `usuario_id` bigint unsigned NOT NULL,
  `evento_id` bigint unsigned NOT NULL,
  `precio_pagado` decimal(10,2) NOT NULL,
  `estado` enum('activa','cancelada','usada','transferida') COLLATE utf8mb4_unicode_ci DEFAULT 'activa',
  `fecha_compra` datetime(3) DEFAULT NULL,
  `fecha_cancelacion` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_entradas_codigo_qr` (`codigo_qr`),
  KEY `idx_entradas_usuario_id` (`usuario_id`),
  KEY `idx_entradas_evento_id` (`evento_id`),
  KEY `idx_entradas_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_eventos_entradas` FOREIGN KEY (`evento_id`) REFERENCES `eventos` (`id`),
  CONSTRAINT `fk_usuarios_entradas` FOREIGN KEY (`usuario_id`) REFERENCES `usuarios` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `entradas`
--

LOCK TABLES `entradas` WRITE;
/*!40000 ALTER TABLE `entradas` DISABLE KEYS */;
/*!40000 ALTER TABLE `entradas` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `eventos`
--

DROP TABLE IF EXISTS `eventos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `eventos` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `titulo` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `descripcion` text COLLATE utf8mb4_unicode_ci,
  `fecha_hora` datetime(3) NOT NULL,
  `duracion_minutos` bigint NOT NULL DEFAULT '120',
  `lugar` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `direccion` varchar(300) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ciudad` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `categoria` enum('musica','deporte','cultura','teatro_cine','conferencia','otro') COLLATE utf8mb4_unicode_ci NOT NULL,
  `capacidad_total` bigint NOT NULL,
  `entradas_vendidas` bigint DEFAULT '0',
  `precio_base` decimal(10,2) NOT NULL,
  `imagen_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `estado` enum('activo','cancelado','agotado','finalizado') COLLATE utf8mb4_unicode_ci DEFAULT 'activo',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_eventos_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `eventos`
--

LOCK TABLES `eventos` WRITE;
/*!40000 ALTER TABLE `eventos` DISABLE KEYS */;
INSERT INTO `eventos` VALUES (1,'Lollapalooza Argentina 2026','El festival de música más grande del país con artistas internacionales.','2026-11-15 14:00:00.000',480,'Hipódromo de Palermo','Av. del Libertador 4101','Buenos Aires','musica',50000,32000,45000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL),(2,'River vs Boca - Superclásico','El partido más esperado del año en el Monumental.','2026-12-01 21:00:00.000',120,'Estadio Monumental','Av. Figueroa Alcorta 7597','Buenos Aires','deporte',84000,70000,25000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL),(3,'TEDx Buenos Aires 2026','Charlas inspiradoras de líderes y pensadores de todo el mundo.','2026-11-20 09:00:00.000',480,'Teatro Gran Rex','Av. Corrientes 857','Buenos Aires','conferencia',1200,400,12000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL),(4,'Carmen - Teatro Colón','La ópera más famosa de Bizet en el teatro más importante de Latinoamérica.','2026-11-25 20:00:00.000',150,'Teatro Colón','Cerrito 628','Buenos Aires','teatro_cine',2500,1800,35000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL),(5,'Festival de Jazz Córdoba','Tres días de jazz en vivo con músicos locales e internacionales.','2026-12-05 18:00:00.000',360,'Plaza San Martín','Plaza San Martín s/n','Córdoba','musica',5000,1200,8000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL),(6,'Maratón Ciudad de Buenos Aires','La maratón anual más popular de Argentina con 42km por la ciudad.','2026-10-10 07:00:00.000',360,'Obelisco','Av. 9 de Julio s/n','Buenos Aires','deporte',20000,18500,5000.00,'','activo','2026-06-16 15:57:00.000','2026-06-16 15:57:00.000',NULL);
/*!40000 ALTER TABLE `eventos` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `usuarios`
--

DROP TABLE IF EXISTS `usuarios`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `usuarios` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nombre` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `apellido` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `rol` enum('cliente','administrador') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'cliente',
  `telefono` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `fecha_registro` datetime(3) DEFAULT NULL,
  `activo` tinyint(1) DEFAULT '1',
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_usuarios_email` (`email`),
  KEY `idx_usuarios_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `usuarios`
--

LOCK TABLES `usuarios` WRITE;
/*!40000 ALTER TABLE `usuarios` DISABLE KEYS */;
/*!40000 ALTER TABLE `usuarios` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-06-16 16:23:04
