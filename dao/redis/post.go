package redis

import (
	"bluebell/models"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func getIDsFormKey(key string, page, size int64) ([]string, error) {
	// 2.确定查询的索引起始点
	start := (page - 1) * size
	end := start + size - 1
	// 3.ZRevRange 查询  （大--->小）
	return client.ZRevRange(key, start, end).Result() // 倒序获取有序集合key从start下标到end下标的元素
}

func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从redis获取id
	// 1.根据用户请求中携带的order参数确定要查询的redis的key
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}

	return getIDsFormKey(key, p.Page, p.Size)

}

// GetPostVoteData 根据ids查询每篇帖子投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	/*	data = make([]int64, 0, len(ids))
		for _, id := range ids {
			key := getRedisKey(KeyPostVotedZSetPF+id)
			// 查找key中分数是1的元素的数量======统计每篇帖子的赞成票的数量
			v := client.ZCount(key, "1", "1").Val()
			data = append(data, v)
		}
	*/

	// 使用pipeline一次发送多条命令，减少RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		pipeline.ZCount(key, "1", "1")
	}
	cmders, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	data = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

// GetCommunityPostIDsInOrder 按社区查找ids
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	}
	// 使用zinterstore把分区的帖子set与帖子分数的zset生成一个新的zset
	// 针对新的zset按之前的逻辑取数据

	// 社区的key
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(p.CommunityID)))
	// 利用缓存key减少zinterstore执行的次数
	key := orderKey + strconv.Itoa(int(p.CommunityID)) // 缓存key
	if client.Exists(key).Val() < 1 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(key, redis.ZStore{
			Aggregate: "MAX",
		}, cKey, orderKey) // 按照post:time/post:score 查   ZInterStore计算
		pipeline.Expire(key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec()
		if err != nil {
			return nil, err
		}
	}
	// 存在的话直接根据key查询ids
	return getIDsFormKey(key, p.Page, p.Size)

}
