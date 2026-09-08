CREATE TABLE IF NOT EXISTS `role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `role_type` varchar(127) NOT NULL DEFAULT 'User',
  `role_name` varchar(127) NOT NULL,
  `enable` enum('Enable','Disable') NOT NULL DEFAULT 'Enable',
  `tenant_id` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_role_deleted_at` (`deleted_at`),
  KEY `idx_role_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
