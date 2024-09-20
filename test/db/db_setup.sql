CREATE TABLE `srd_routes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `origin` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'The origin navaid or airport for the route',
  `destination` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'The destination navaid or airport for the route',
  `minimum_level` int DEFAULT NULL COMMENT 'The minimum flight level for the route',
  `maximum_level` int NOT NULL COMMENT 'The maximum flight level for the route',
  `route_segment` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'The route segment',
  `sid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'The SID used at the start of the route',
  `star` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'The STAR used at the end of the route',
  PRIMARY KEY (`id`),
  KEY `srd_routes_origin_index` (`origin`),
  KEY `srd_routes_destination_index` (`destination`),
  KEY `srd_routes_minimum_level_index` (`minimum_level`),
  KEY `srd_routes_maximum_level_index` (`maximum_level`)
) ENGINE=InnoDB AUTO_INCREMENT=166866 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `srd_notes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `note_text` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=514 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `srd_note_srd_route` (
  `srd_route_id` bigint unsigned NOT NULL,
  `srd_note_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`srd_note_id`,`srd_route_id`) USING BTREE,
  KEY `srd_note_srd_route_srd_route_id_foreign` (`srd_route_id`) USING BTREE,
  CONSTRAINT `srd_note_srd_route_srd_note_id_foreign` FOREIGN KEY (`srd_note_id`) REFERENCES `srd_notes` (`id`) ON DELETE CASCADE,
  CONSTRAINT `srd_note_srd_route_srd_route_id_foreign` FOREIGN KEY (`srd_route_id`) REFERENCES `srd_routes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
