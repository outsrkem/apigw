# apigw

## api gateway

---

- 数据库操作

```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS apigw CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER 'apigw'@'%' identified by '123456';

-- 授权数据库权限给用户
GRANT ALL ON apigw.* TO 'apigw'@'%';
```

- 加密数据库密码

```shell
docker run -it --rm swr.cn-north-1.myhuaweicloud.com/onge/apigw:build_0 apigw -encrypt 'password'
```

- 构建运行指南

```sh
./.cid/build.sh
./output/apigw -f apigw.yaml
```
