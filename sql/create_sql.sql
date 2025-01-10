DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `phone_number` varchar(20)     NOT NULL COMMENT '手机号',
    `nickname`     varchar(20)     NOT NULL COMMENT '昵称',
    `password`     varchar(255)    NOT NULL COMMENT '密码',
    `create_time`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone_number` (`phone_number`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `message`;
CREATE TABLE `message`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `user_id`      bigint unsigned NOT NULL COMMENT '发送者用户id',
    `group_id`    bigint unsigned NOT NULL COMMENT '房间id',
    `content`      blob            NOT NULL COMMENT '消息内容',
    `create_time`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `group`;
CREATE TABLE `group`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `name`        varchar(50)     NOT NULL COMMENT '群组名称',
    `owner_id`    bigint unsigned NOT NULL COMMENT '群主id',
    `create_time` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

DROP TABLE IF EXISTS `group_user`;
CREATE TABLE `group_user`
(
    `id`          bigint  unsigned  NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `group_id`    bigint  unsigned  NOT NULL COMMENT '群组id',
    `user_id`     bigint  unsigned  NOT NULL COMMENT '用户id',
    `group_type`  tinyint unsigned  NOT NULL COMMENT '房间类型',
    `create_time` datetime          NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime          NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_group_id_user_id` (`group_id`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;