/*
 Source Server Type    : MySQL
 Source Server Version : 50744
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for api_domain
-- ----------------------------
DROP TABLE IF EXISTS `api_domain`;
CREATE TABLE `api_domain`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `group_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `ssl_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'SSL状态（0-关闭，1-开启）',
  `ssl_cert` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT 'SSL证书内容',
  `ssl_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT 'SSL私钥内容',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL,
  `update_time` bigint(20) NOT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_domain_name`(`name`) USING BTREE,
  INDEX `group_id`(`group_id`) USING BTREE,
  CONSTRAINT `api_domain_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `api_group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for api_group
-- ----------------------------
DROP TABLE IF EXISTS `api_group`;
CREATE TABLE `api_group`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `remark` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '0-禁用,1-启用',
  `create_time` bigint(20) NOT NULL,
  `update_time` bigint(20) NOT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_group_name`(`name`) USING BTREE,
  INDEX `id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for api_interface
-- ----------------------------
DROP TABLE IF EXISTS `api_interface`;
CREATE TABLE `api_interface`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `group_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `protocol` enum('HTTP','HTTPS') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端协议',
  `method` enum('GET','POST','DELETE','PATCH','PUT','HEAD','OPTIONS','Any') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'HTTP方法',
  `req_uri` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'API路径',
  `backend_uri` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端api',
  `auth` enum('NONE','UIAS','TOKEN') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '认证类型',
  `mode` enum('EXACT','PREFIX') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'API匹配方式',
  `lc_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '关联负载通道ID',
  `rate_limit` int(11) NOT NULL DEFAULT 0 COMMENT '接口限流（QPS，0-不限流）',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `publish` tinyint(4) NOT NULL DEFAULT 0 COMMENT '发布状态（0-未发布，1-测试中，2-已发布，3-已下线）',
  `create_time` bigint(20) NOT NULL,
  `update_time` bigint(20) NOT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_api_path_method`(`req_uri`, `method`) USING BTREE,
  UNIQUE INDEX `uk_group_id_name`(`group_id`, `name`) USING BTREE,
  INDEX `idx_group_id`(`group_id`) USING BTREE,
  INDEX `idx_lb_id`(`lc_id`) USING BTREE,
  CONSTRAINT `api_interface_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `api_group` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `api_interface_ibfk_2` FOREIGN KEY (`lc_id`) REFERENCES `load_channel` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for lc_ca
-- ----------------------------
DROP TABLE IF EXISTS `lc_ca`;
CREATE TABLE `lc_ca`  (
  `kid` bigint(11) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `CN` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `NotBefore` bigint(20) NOT NULL DEFAULT 0,
  `NotAfter` bigint(20) NULL DEFAULT 0,
  `cert` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  `create_time` bigint(20) NOT NULL,
  `update_time` bigint(20) NOT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  INDEX `id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for load_channel
-- ----------------------------
DROP TABLE IF EXISTS `load_channel`;
CREATE TABLE `load_channel`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `backend` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端地址列表,逗号分隔',
  `ca_cert` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `timeout` int(11) NOT NULL DEFAULT 3000 COMMENT '后端超时时间（毫秒）',
  `hcinterval` int(11) NULL DEFAULT 5000 COMMENT '健康检查间隔（毫秒）',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL,
  `update_time` bigint(20) NOT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_lb_name`(`name`) USING BTREE,
  INDEX `id`(`id`) USING BTREE,
  INDEX `ca_cert`(`ca_cert`) USING BTREE,
  INDEX `idx_lb_id_status`(`kid`, `status`) USING BTREE,
  CONSTRAINT `load_channel_ibfk_1` FOREIGN KEY (`ca_cert`) REFERENCES `lc_ca` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1000 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;
