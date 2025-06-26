# Online Boutique

Online Boutique 是 Google 开源的微服务系统，其具体介绍可以去官方仓库查看。

> https://github.com/GoogleCloudPlatform/microservices-demo/tree/main

## 文件架构

```text
├── docs
├── helm-chart：使用 helm 安装的配置文件
│   ├── Chart.yaml
│   ├── README.md
│   ├── templates
│   └── values.yaml
├── kubernetes：使用 kubectl 安装的配置文件
│   ├── loadgenerator.yaml
│   └── onlineboutique
├── make-docker-images.sh：构建微服务镜像的脚本
├── mysql：数据库数据备份
│   ├── addatabase.sql
│   └── productdb.sql
├── README.md
└── src：微服务源代码
    ├── adservice
    ├── cartservice
    ├── checkoutservice
    ├── currencyservice
    ├── emailservice
    ├── frontend
    ├── loadgenerator
    ├── locustexporter
    ├── paymentservice
    ├── productcatalogservice
    ├── recommendationservice
    └── shippingservice
```

## 微服务系统部署

### 1. 环境准备

微服务系统可以部署在集群环境，或者是以容器运行在本地 docker 环境。请确保安装对应的工具。

### 2. 构建镜像

`src/` 文件夹下存放的是每个微服务的源代码，且都提供了对应的 `Dockerfile`。

其中 `loadgenerator` 是流量注入的组件， `locustexporter` 是采集流量指标的组件。

批量创建镜像，执行命令：`bash make-docker-images.sh`

> 请修改 make-docker-images.sh 中的repository和tag

### 3. 前置依赖

#### 3.1 数据库

数据库使用的是 `mysql:5.7`。

`mysql/` 里是 `productcatalogservice` `adservice` 微服务使用的数据库的数据备份。

首先需要创建 productdb 和 addatabase 两个数据库。

```bash
CREATE DATABASE productdb;
CREATE DATABASE addatabase;
```

然后将数据备份导入数据库。

```bash
mysql -h <HOST> -P <PORT> -u <USER> -p productdb < productdb.sql
mysql -h <HOST> -P <PORT> -u <USER> -p addatabase < addatabase.sql
```

#### 3.2 Jaeger

微服务源码中已经添加了 `jaeger` 相关的上报点，因此部署微服务系统前必须先安装好 `jaeger`。

### 4. 部署安装

#### 4.1 使用 helm

`helm-chart/` 是使用 helm 部署微服务系统的配置文件。 `templates` 下是每个微服务具体的配置文件，需要将其中关于 `jaeger` 相关的部分，和数据库相关的部分，修改为当前环境下的配置。`values.yaml` 文件需要更改镜像的仓库和tag信息。

更改完后，在根目录下执行安装命令：`helm install YOUR_NAME -n YOUR_NAMESPACE ./helm-chart`。

卸载，执行命令：`helm uninstall YOUR_NAME -n YOUR_NAMESPACE`。

#### 4.2 使用 kubectl

`kubernetes/` 下存放的是每个微服务系统单独的配置文化，同样需要修改对应的内容。

部署微服务系统，在 `onlineboutique/` 下执行：`kubectl apply -f . -n YOUR_NAMESPACE`。

启动流量注入：`kubectl apply -f loadgenerator.yaml -n YOUR_NAMESPACE`。

卸载，执行 `kubectl delete -f YOUR_YAML_FILE -n YOUR_NAMESPACE`。

### 5. 流量说明

本项目提供的是 `Locust` 流量注入工具，具体详细可以去官方文档查看，`Locust` 使用的注入文件在 `src/loadgenerator` 中给出。

> 为方便进行压测，本项目使用locust-exporter访问locust的8089端口
> 并且需要在prometheus中添加对应的scrape-job
> 示例如下
> ```
> - job_name: 'locust-exporter-k8s'
>   kubernetes_sd_configs:
>       - role: pod
>   relabel_configs:
>       - source_labels: [__meta_kubernetes_pod_label_component]
>       regex: locust-metrics
>       action: keep
>       - source_labels: [__meta_kubernetes_pod_ip]
>       target_label: __address__
>       replacement: $1:9646
>       - source_labels: [__meta_kubernetes_namespace]
>       target_label: namespace
> ```
> 想要采集对应的locust容器，就在locust的配置文件labels中添加"component: locust-metrics"
