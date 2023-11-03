# grpc/gin 项目协同管理系统
- **应用技术**：golang,gin,gorm,grpc,redis,mysql,zap,etcd,jwt,minio,jaeger,docker,ELK(elasticsearch,logstash,Kibana),kafka
- **项目描述**：新一代的多项目管理系统，具有项目流程模版，适配多种项目类型，助力企业精细化管理，便携的任务看板助力效率提升，支持任务文件上传和成员的权限控制

- **技术亮点**：
1. 项目分为api, project和user模块设计，gin框架实现api路由分发，grpc 实现 user/project模块服务调用
2. 采用grpc拦截器实现redis统一缓存，并且利用kafka消息队列的方式解决缓存和数据库的一致性问题
3. 采用etcd实现服务注册和服务发现，jaeger实现请求的服务链路的可视化，nacos实现配置文件的在线更改和自动更新，ELK 搭配Kafka 实现日志采集，独立的速率限制组件限制单位时间内的请求数
4. 采用mysql主从架构实现读写分离，在业务代码层面实现事务来保证数据的一致性
5. 采用minIO实现文件的单独和分片上传至对象存储中心，利用docker-compose完成数据库和中间件的部署

- **个人收获**：了解grpc 相关微服务 项目流程。


## 使用说明

1. 运行 build.bat => 选择 2 编译Linux（如果 api/project/user 模块下的 target文件夹有相应文件，则不需要重新编译）
2. 运行 run.bat => docker 会自动编排容器运行
3. 同步运行库内前端文件 pearProject，终端运行 npm run serve

## 各中间件设置相关网页

1. nacos：
   * http://127.0.0.1:8848/nacos/index.html
   * Account：nacos
   * Password：nacos
2. Kafka:
   * http://localhost:9000/
3. elastic(ELK):
   * http://localhost:5601/
4. jaeger:
   * http://localhost:16686/search
5. minIO：
    * http://localhost:9001/login
    * account: admin
    * password: admin123456
6. 缓存一致性：
   * https://www.dtm.pub/app/cache.html

   