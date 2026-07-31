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
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `group_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '分组id',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '域名（如api.user.com）',
  `ssl_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT 'SSL状态（0-关闭，1-开启）',
  `ssl_cert` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT 'SSL证书内容',
  `ssl_key` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT 'SSL私钥内容',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL COMMENT '创建时间戳(秒)',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间戳(秒)',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_domain_name`(`name`) USING BTREE COMMENT '域名唯一约束',
  INDEX `group_id`(`group_id`) USING BTREE,
  CONSTRAINT `api_domain_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `api_group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '域名配置表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for api_group
-- ----------------------------
DROP TABLE IF EXISTS `api_group`;
CREATE TABLE `api_group`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'uud',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '分组名称（如用户服务）',
  `remark` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '分组描述',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL COMMENT '创建时间戳(秒)',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间戳(秒)',
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_group_name`(`name`) USING BTREE COMMENT '分组名称唯一约束',
  INDEX `id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'API分组表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for api_interface
-- ----------------------------
DROP TABLE IF EXISTS `api_interface`;
CREATE TABLE `api_interface`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'API ID',
  `group_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '所属分组ID',
  `protocol` enum('HTTP','HTTPS') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端协议',
  `method` enum('GET','POST','DELETE','PATCH','PUT','HEAD','OPTIONS','Any') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'HTTP方法（如GET,POST）',
  `req_uri` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'API路径（如/api/user/login）包含请求参数，用{}标识',
  `backend_uri` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端api（如将/api/v1/*重写为/*）',
  `auth` enum('NONE','UIAS','TOKEN') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '认证类型',
  `mode` enum('EXACT','PREFIX') CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'API的匹配方式prefix：前缀匹配,exact：精确匹配',
  `lc_id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '关联负载通道ID',
  `rate_limit` int(11) NOT NULL DEFAULT 0 COMMENT '接口限流（QPS，0-不限流）',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL COMMENT '创建时间戳(秒)',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间戳(秒)',
  `publish_status` tinyint(4) NOT NULL DEFAULT 0 COMMENT '发布状态（0-未发布，1-测试中，2-已发布，3-已下线）',
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_api_path_method`(`req_uri`, `method`) USING BTREE COMMENT '路径+方法唯一约束（避免重复API）',
  INDEX `idx_group_id`(`group_id`) USING BTREE COMMENT '分组查询索引',
  INDEX `idx_lb_id`(`lc_id`) USING BTREE COMMENT '负载通道关联索引',
  CONSTRAINT `api_interface_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `api_group` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `api_interface_ibfk_2` FOREIGN KEY (`lc_id`) REFERENCES `load_channel` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 12 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'API详情表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Table structure for lc_ca
-- ----------------------------
DROP TABLE IF EXISTS `lc_ca`;
CREATE TABLE `lc_ca`  (
  `kid` bigint(11) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '证书ID',
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '证书CN',
  `cert` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT 'base64编码证书',
  `create_time` bigint(20) NULL DEFAULT NULL,
  `update_time` bigint(20) NULL DEFAULT NULL,
  PRIMARY KEY (`kid`) USING BTREE,
  INDEX `id`(`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for load_channel
-- ----------------------------
DROP TABLE IF EXISTS `load_channel`;
CREATE TABLE `load_channel`  (
  `kid` bigint(20) NOT NULL AUTO_INCREMENT,
  `id` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '负载通道ID',
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '通道名称（如用户服务集群）',
  `backend` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '后端地址列表（逗号分隔，如192.168.1.1:8080,192.168.1.2:8080）',
  `ca_cert` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '后端ca证书',
  `timeout` int(11) NOT NULL DEFAULT 3000 COMMENT '后端超时时间（毫秒）',
  `hcinterval` int(11) NULL DEFAULT 5000 COMMENT '健康检查间隔（毫秒）health_check_interval',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
  `create_time` bigint(20) NOT NULL COMMENT '创建时间戳(秒)',
  `update_time` bigint(20) NOT NULL COMMENT '更新时间戳(秒)',
  PRIMARY KEY (`kid`) USING BTREE,
  UNIQUE INDEX `uk_lb_name`(`name`) USING BTREE COMMENT '负载通道名称唯一约束',
  INDEX `idx_lb_id_status`(`kid`, `status`) USING BTREE COMMENT 'ID+状态查询索引',
  INDEX `id`(`id`) USING BTREE,
  INDEX `ca_cert`(`ca_cert`) USING BTREE,
  CONSTRAINT `load_channel_ibfk_1` FOREIGN KEY (`ca_cert`) REFERENCES `lc_ca` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '负载通道表' ROW_FORMAT = DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;
