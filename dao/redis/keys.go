package redis

// redis key
// redis key 注意使用命名空间的方式，方便查询与区分

const (
	Prefix             = "bluebell:"
	KeyPostTimeZSet    = "post:time"  //ZSet 帖子及发帖时间
	KeyPostScoreZSet   = "post:score" //ZSet 帖子及投票分数
	KeyPostVotedZSetPF = "post:voted" //ZSet 记录用户及投票类型，参数是post_id，PF不完整的key
	KeyCommunitySetPF  = "community:" //set，保存每个分区内帖子的id
)

// getRedisKey 给redis key 加上前缀
func getRedisKey(key string) string {
	return Prefix + key
}
