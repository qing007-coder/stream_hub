### 未来要做的

#### 日志
 ##### 日志做重试和本地持久化处理 可以加上采样的处理


#### 任务调度器
##### 可以加上优先级 然后可以加上任务队列的优先级 然后可以把失败的任务放入死信队列
#####  worker的taskhealth加上分布式锁  黑名单用集合来存 

#### 互动模块
##### 可以用middleware来消息推送 异步发送任务发送修改user的点赞数还有一些别的字段 like follow favourite应该异步落库 后期再更改

### cache
#### 写回法

### 可以写在简历上的 第一遍用的是所有的worker作为payload但是发现worker数量上来了 payload会很大 janitor容易发生竞态
### 第二版用的是死亡的worker作为payload 但是如果janitor未扫描到就被覆盖了 janitor就清理不了死掉的worker队列了
### 第三版是node发心跳（无payload） worker死亡后用LPush去报丧 让janitor清理