### ec2配置多主集群
#### 启机器可以用docker-machine
>- docker-machine create --driver amazonec2 --amazonec2-ami ami-0a22b8776bb32836b  --amazonec2-ssh-user ubuntu --amazonec2-endpoint ec2.cn-northwest-1.amazonaws.com.cn --amazonec2-vpc-id vpc-034b2afd5c458e2d1 --amazonec2-subnet-id subnet-0aac189f8f33c7da9 --amazonec2-zone cn-northwest-1a --amazonec2-keypair-name taobaoke --amazonec2-ssh-keypath /c/Users/Lan/.ssh/id_rsa   --amazonec2-open-port 8000 --amazonec2-region cn-northwest-1  node1  
#### 默认是ubuntu用户,为了集群安装方便需要root登录copy-id
>- su root
>- vim /etc/ssh/sshd_config
> >PasswordAuthentication yes
> >PermitRootLogin yes
>- systemctl restart sshd
[更新ec2 ubuntu18.04源](https://blog.csdn.net/qq_41433316/article/details/107802880)

[kubeasz](https://github.com/easzlab/kubeasz)

#### kubeasz 用的traefik是v1.x版本的 过时了
[traefik2.2](https://docs.traefik.io/user-guides/crd-acme/)

#### 存储卷用的openebs
[openebs](https://github.com/openebs/openebs)

#### k8s部署Replica Set mongo 
[mongo](https://mp.weixin.qq.com/s/n0tpqBSnZ9x7jCeKFu4IAQ)

#### k8s部署 redis cluster
[redis](https://github.com/zuxqoj/kubernetes-redis-cluster)
[readme](https://www.jianshu.com/p/65c4baadf5d9)
[cluster 脚本](https://redis.io/topics/cluster-tutorial/)

#### k8s部署 bitnami/redis
[redis](https://github.com/bitnami/charts/tree/master/bitnami/redis)