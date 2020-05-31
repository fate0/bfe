# BFE with keyless

* [BFE](https://github.com/baidu/bfe) is an open-source layer 7 load balancer derived from proprietary Baidu FrontEnd.
* [gokeyless](https://github.com/cloudflare/gokeyless) is an implementation Cloudflare's Keyless SSL Protocol in Go. 

## 介绍

BFE with keyless 是基于 BFE v0.10.0 版本进行修改，主要合并了两个 crypto/tls 官方 commit，
以便于 bfe_tls 能够不直接使用 Go 代码中的 ECDSA/RSA 的 sign/decrypt 操作，转而使用用户提供的 sign/decrypt 操作。
增加了 `mod_keyless` 模块，将 gokeyless 应用到 BFE 上。

## 开始

1. 安装 BFE with keyless 和 gokeyless server
```shell script
git clone https://github.com/fate0/bfe.git
make
git clone https://github.com/cloudflare/gokeyless
make && make install
```

默认 gokeyless 的配置目录为 /etc/keyless/

2. 生成证书
```shell script
cd output/conf/mod_keyless
python gencert.py
``` 

3. 配置 gokeyless server

将第二步生成的 `server.pem`, `server-key.pem`, `key` 移动到 gokeyless 配置目录中，将 `ca.pem` 拷贝到 gokeyless 配置目录中
```shell script
mv server.pem server-key.pem /etc/keyless/
mv key/* /etc/keyless/
cp ca.pem /etc/keyless/
```

删除 auth_csr 配置，将 cloudflare_ca_cert 配置成 ca.pem 的路径，`service gokeyless restart` 重启 gokeyless

4. 配置 BFE

更改 BFE 所在机器的 /etc/hosts 文件，将 `www.keyless.com` 指向 gokeyless server 机器 ip
```shell script
echo 127.0.0.1 www.keyless.com >> /etc/hosts
```

剩下就是根据后端情况对 BFE 进行配置了，因为之前生成的是 CN 为 `www.[0-9].com` 的证书(偶数域名为 ecdsa 证书，奇数域名为 rsa 证书)，
所以需要将 `www.[0-9].com` 的域名加到 host_rule.data 中，具体的配置信息可以看 BFE 的文档

## 测试

在发起请求的机器上绑定 `www.[0-9].com` 对应的 BFE ip，执行 

```shell script
curl https://www.1.com:8443 -k
curl --cacert ca.pem https://www.1.com:8443
```

如果能正常访问，表示整个 keyless 设计跑通了

## 压力测试

使用 gokeyless 中 bench 程序对 keyless server 进行压力测试：


#### ecdsa + bandwidth

```shell script
./bench -bandwidth -cert client.pem  -key client-key.pem  -ca ca.pem  -server www.keyless.com -ski 0825cecb524cc1d72dbb174398b5dac491da80ec -sni www.2.com -workers 32 -gmp 32 -histogram-max=30ms -histogram-step=1ms
```

1. 本地调用
```
Total operations completed: 387210
Average operation duration: 25.825µs

Total operations completed: 380946
Average operation duration: 26.25µs

Total operations completed: 404601
Average operation duration: 24.715µs
```

2. 远程调用(同机房)
```
Total operations completed: 380275
Average operation duration: 26.296µs

Total operations completed: 383653
Average operation duration: 26.065µs

Total operations completed: 397747
Average operation duration: 25.141µs
```

可以看到本地和远程的结果是相差无几的

#### rsa + bandwidth

```shell script
./bench -bandwidth -cert client.pem  -key client-key.pem  -ca ca.pem  -server www.keyless.com -ski 6d5dd5cd9054726ea8afe8ada2462fe38fc7a3e1 -sni www.3.com -workers 32 -gmp 32 -histogram-max=10ms -histogram-step=0.5ms
```

1. 本地调用
```
Total operations completed: 66959
Average operation duration: 149.345µs

Total operations completed: 67628
Average operation duration: 147.867µs

Total operations completed: 65786
Average operation duration: 152.008µs
```

2. 远程调用
```
Total operations completed: 68998
Average operation duration: 144.931µs

Total operations completed: 64558
Average operation duration: 154.899µs

Total operations completed: 65906
Average operation duration: 151.731µs
```

可以看到本地和远程的结果是相差无几的，但是 ecdsa 的 sign 速度明显比 rsa 要快

#### ecdsa + latency

```shell script
./bench -pause 1ms -cert client.pem  -key client-key.pem  -ca ca.pem  -server www.keyless.com -ski 115c176f6e989d805ed56605db37254a48887021 -sni www.4.com -workers 32 -gmp 32 -histogram-max=2ms -histogram-step=0.2ms
```

```
[    0s,  200µs)  37116 | ===============
[ 200µs,  400µs) 169037 | ======================================================================
[ 400µs,  600µs)  19887 | ========
[ 600µs,  800µs)   4388 | =
[ 800µs,    1ms)   1249 | 
[   1ms,  1.2ms)    747 | 
[ 1.2ms,  1.4ms)    561 | 
[ 1.4ms,  1.6ms)    450 | 
[   2ms,      -)   1726 |
```

绝大多数 sign/decrypt 请求都能在 1ms 内返回

### rsa + latency

```shell script
./bench -pause 1ms -cert client.pem  -key client-key.pem  -ca ca.pem  -server www.keyless.com -ski aa8637f354ec54175b443e8c9063bf2ebb9e8094 -sni www.5.com -workers 32 -gmp 32 -histogram-max=10ms -histogram-step=0.5ms
```

```
[  2ms, 2.5ms)  2391 | ===========
[2.5ms,   3ms)  8523 | =======================================
[  3ms, 3.5ms)  5350 | ========================
[3.5ms,   4ms)  9854 | =============================================
[  4ms, 4.5ms) 15085 | ======================================================================
[4.5ms,   5ms)  6304 | =============================
[  5ms, 5.5ms)  3546 | ================
[5.5ms,   6ms)  1940 | =========
[  6ms, 6.5ms)  1189 | =====
[6.5ms,   7ms)   818 | ===
[  7ms, 7.5ms)   651 | ===
[7.5ms,   8ms)   601 | ==
[  8ms, 8.5ms)   404 | =
[8.5ms,   9ms)   339 | =
[ 10ms,     -)  1587 | =======
```

所有 sign/decrypt 请求都超过了 1ms，而且在观察 cpu 利用率时，发现 rsa 的 cpu 利用率比 ecdsa 要高不少


### ecdsa + siege
```shell script
siege -d 0 -c 100 -t 30S https://www.2.com:8443
```

```
Lifting the server siege...
Transactions:		       88228 hits
Availability:		      100.00 %
Elapsed time:		       29.94 secs
Data transferred:	      330.85 MB
Response time:		        0.03 secs
Transaction rate:	     2946.83 trans/sec
Throughput:		       11.05 MB/sec
Concurrency:		       94.75
Successful transactions:       88230
Failed transactions:	           0
Longest transaction:	        3.66
Shortest transaction:	        0.00
```

#### rsa + siege

```shell script
siege -d 0 -c 100 -t 30S https://www.3.com:8443
```

```
Lifting the server siege...
Transactions:		       77984 hits
Availability:		      100.00 %
Elapsed time:		       29.09 secs
Data transferred:	      292.44 MB
Response time:		        0.04 secs
Transaction rate:	     2680.78 trans/sec
Throughput:		       10.05 MB/sec
Concurrency:		       95.31
Successful transactions:       77984
Failed transactions:	           0
Longest transaction:	        3.57
Shortest transaction:	        0.00
```

## 问题

从刚刚 bandwidth 测试可以看出 rsa 每秒只能处理 2k 多个 sign/decrypt 请求，一旦超过这个数据，就会出现问题，
而且如果 client 设置超时重试，那问题会变得更严重，不停的超时重试只会让 keyless server 处理越来越多的请求，直至崩溃。
最后 keyless client(BFE) 发出去的 sign/decrypt 不会随着 client (browser) 中断而中断，等于就是让 keyless server 做无用功。 


## 代码比较

https://github.com/baidu/bfe/compare/v0.10.0...fate0:keyless

