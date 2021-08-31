# 火山公有云密码重置

## 1. 实现逻辑：
  此agent是安装在火山云虚拟机内的，通过此agent提供虚拟机内的重置认证功能，此服务仅在开机后启动一次，之后将不再运行，客户重置认证需要重启虚拟机

## 2. 对metadata请求接口
  2.1 http://100.96.0.96/volcstack/latest/instance_id (GET)
  获取虚拟机的ID信息，使用来校验可用的数据源接口

  2.2 http://100.96.0.96/volcstack/latest/reset_authentication_version (POST)
  推送agent自身的version，metadata会根据虚拟机的version信息判断虚拟机是否支持重置功能，未推送将不再支持此功能

  2.3 http://100.96.0.96/volcstack/latest/reset_password （GET）
  获取要重置的虚拟机密码，此密码会在用户浏览器进行加密，此得到的结果会是一个加密后的密码

  2.4 /volcstack/latest/reset_del_pubkey (GET)
  获取要重置的虚拟机的预备删除key，此结果会删除 /root/.ssh/authorized_keys 其中的公钥

  2.5 /volcstack/latest/reset_add_pubkey (GET)
  获取要重置的虚拟机的预备增加key，此结果会增加 /root/.ssh/authorized_keys 其中的公钥
