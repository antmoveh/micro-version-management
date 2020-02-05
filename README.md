# micro-version-management

##### 背景
 - 随着业务功能拆分,微服务组件逐渐增多，各个服务的最终版本逐渐不一,而且最终可用release版镜像地址存储于gitlab的升级部署issue中，更新镜像后需要同步更新升级文档，这个过程繁琐且容易出错。
 
##### 问题
 - 微服务增多，各个服务的最终镜像版本不一致
 - 确定最终各组件release可用版本镜像，需要查看gitlab升级issue才能确定
 - 镜像迁移，即从repository.****.com迁移到repository.xxx.com
 
##### 解决方案
 - 定义release版本镜像名规则`project/release/+组件名称`，示例`project/release/cluster-management`
 - 定义release版本镜像tag规则为`大版本号-编译序号`，示例`v1.9-10`
 - 各个组件模板放置在目录A下文件名称需与镜像名称相同，示例`cluster-management`
 - 如此我们便可定义v1.9-xx最后一个版本号即为release最新版本
 - 执行过程:先遍历yaml所在目录-->得到镜像名称-->通过镜像名称搜索所有符合条件的镜像tag-->计算出最后一个镜像tag，生成可执行的yaml文件
 >- 备注
 >- 之所以采用在镜像名称中增加release字段而非在imageTag中增加release，主要原因如下
 >- 1.流水线中镜像名称可以随意定制而tag是自动生成的修改不易
 >- 2.docker search是不支持搜索镜像所有tag的，我们是通过rest api的方式搜索镜像下所有tag，image name中加入release字段可有效减少搜索范围

##### 技术调研
  >- 对于接收命令行参数、读写文件等不需要额外关注，主要的关注点在于docker client端不支持列出远程仓库中该镜像下所有tag，所以调研重点放在了这里，结果如下
  
  - 期望效果 `exe v1.9-->即生成yaml文件`
  - 根据镜像搜索所有tag
    - docker hub:
    ```
    https://registry.hub.docker.com/v2/repositories/library/debian/tags?page_size=20&&page=1
    ```
    - nexus:
    ```
    login: curl -d "username=YWRtaW4%3D&password=YWRtaW4xMjM%3D" http://10.200.64.38:8081/service/rapture/session
    - cookie中需包含Cookie:NXSESSIONID=e0d078bf-6a16-43f3-9ae9-dea809e22fc3
    tagList: http://repository.***.com/service/rest/v1/search?docker.imageName=moebius/core-api
    ```
    
    - harbor(basicAuth):
    ```
    - 请求方法：req, err := http.NewRequest("GET", url, nil) req.SetBasicAuth("admin", "harbor12345") resp, err := httpClient.Do(req)
    tagList: https://repository.service.cloud.com:8444/api/repositories/moebius/zentao/tags?detail=false
    ```
    
##### 技术选型

  - 开始时，首先想到的是写一段shell脚本完成，可以随着思考的深入感觉逻辑复杂，使用shell脚本需要处理很多业务逻辑且依赖服务器安装jq库才能解析镜像tag列表，当然使用python编写简单快速，但是也是有个致命依赖项问题，需要你的服务器安装对应的python版本及安装python对应版本依赖包；最终决定使用golang实现功能，就像设想的一样编译成一个二进制文件，传入指定镜像仓库、大版本号，便生成最新版本的部署yaml，甚至于自动执行`kubectl apply -f ` 完成了升级部署动作
  
  