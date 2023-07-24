# Mao Service Discovery
|Category|Job|
|---|---|
|Build|[![Go](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/go.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/go.yml) [![Go Static](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/go_static.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/go_static.yml) |
|Test|[![CodeQL](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/codeql-analysis.yml) [![Docker Image CI](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/docker-image.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/docker-image.yml) |
|Binary Release|[![Docker Image Publish](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/docker-publish.yml) [![Node.js Package Publish Binary (Commit)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/npm-publish.yml/badge.svg)](https://github.com/MaoJianwei/Mao_Service_Discovery/actions/workflows/npm-publish.yml) |
|Binary Link|[Github Docker Image](https://github.com/MaoJianwei/Mao_Service_Discovery/pkgs/container/mao_service_discovery) = [NPM official](https://www.npmjs.com/package/mao-service-discovery?activeTab=versions) = [Taobao & Alibaba Cloud mirror](https://npmmirror.com/package/mao-service-discovery)|

## Software Architecture
Please refer to [MODULES.md](https://github.com/MaoJianwei/Mao_Service_Discovery/blob/master/MODULES.md) file.

## Todo List
Please refer to the [agile board](https://github.com/users/MaoJianwei/projects/3).

## Initial need
Discover your service by two methods:

1. Client-Server mode, using gRPC stream.
2. Server-only mode, using ICMP.
3. Using etcd.

### Product: Client-Server mode, using gRPC stream.
#### 1. REST API (JSON format)
![2-json-format.png](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/2-json-format.png)

#### 2. Web Monitor
![2-readable-format.png](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/2-readable-format.png)

#### 3. CLI Output
![2-cli-output.png](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/2-cli-output.png)

#### 4. CLI Parameters
![2-cli-parameters.png](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/2-cli-parameters.png)

### Demo 1: Client-Server mode, using gRPC stream.
![Client-Server mode, using gRPC stream. 1](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/client-server-mode-1.png)

![Client-Server mode, using gRPC stream. 2](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/client-server-mode-2.png)

![client-server-mode-production.png](https://raw.githubusercontent.com/MaoJianwei/MaoServiceDiscovery/master/screenshot/client-server-mode-production.png)

### Demo 2: Using etcd.
![Using etcd.](https://raw.githubusercontent.com/MaoJianwei/Mao_Service_Discovery/master/screenshot/show_using_etcd.png)

## Architect

Jianwei Mao

https://www.MaoJianwei.com/

E-mail: maojianwei2012@126.com

.

![JetBrains Logo](https://account.jetbrains.com/static/favicon.ico) Supported by [JetBrains IDEA Open Source License](https://www.jetbrains.com/?from=Mao_Service_Framework) 2020-2023. 
