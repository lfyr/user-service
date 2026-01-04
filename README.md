安装ETCD
docker run -d  --name etcd-single-node   --restart always  -p 2379:2379  -p 2380:2380  -e ALLOW_NONE_AUTHENTICATION=yes  -e ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379  -v /data/etcd:/bitnami/etcd   bitnami/etcd:latest

编译文件
protoc  --micro_out=. --go_out=. .\proto\user\user.service.proto  