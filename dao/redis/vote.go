package redis

import (
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

/* 简化版投票分数
 用户投一票：+432   86400/200 -> 需要200张赞成票可以给帖子续一天（时间戳+1天）

投票的几种情况：
direction=1
	1. 之前没有投过票，现在投赞成票     --> 更新分数和投票纪录  差值绝对值：1 +432
	2. 之前投反对票，现在改投赞成票   --> 更新分数和投票纪录  差值绝对值：2 +432*2   既要把之前反对的-432加回来，又要再加赞成票的432

direction=0
	1. 之前投过反对票，现在要取消投投票 --> 更新分数和投票纪录  差值绝对值：1 +432
	2. 之前投过赞成票，现在要取消投票   --> 更新分数和投票纪录  差值绝对值：1 -432

direction=-1
	1. 之前没有投过票，现在投反对票 --> 更新分数和投票纪录  差值绝对值：1 -432
	2. 之前投赞成票，现在改投反对票 --> 更新分数和投票纪录  差值绝对值：2 -432*2

投票限制：
每个帖子自发表之日起，一个星期之内允许用户投票
	1. 到期之后将redis中保存的赞成票数与反对票数存储到mysql表中
	2. 到期之后删除 KeyPostVotedZSetPF
*/

const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 // 每一票所占分值
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeated   = errors.New("不允许重复投票")
)

func CreatePost(postID, communityID int64) error {

	pipeline := client.TxPipeline()
	// 帖子时间
	pipeline.ZAdd(getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 帖子分数
	pipeline.ZAdd(getRedisKey(KeyPostScoreZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	// 把帖子id加到社区的set里
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(cKey, postID)
	_, err := pipeline.Exec()
	return err
}

func VoteForPost(userID, postID string, value float64) error {
	//1.判断投票限制
	// 去redis取帖子发布时间  ZScore：返回有序集合key中成员的分数值
	postTime := client.ZScore(getRedisKey(KeyPostTimeZSet), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}
	// 2和3需要放到一个pipeline中

	// 2.更新帖子分数
	// 先查当前用户给当前帖子的投票纪录 1 0 -1
	ov := client.ZScore(getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()
	// 如果这一次投票的值和之前保存的值一致，就不允许投票
	if value == ov {
		return ErrVoteRepeated
	}
	var op float64
	if value > ov {
		op = 1
	} else {
		op = -1
	}
	diff := math.Abs(ov - value) // 计算两次投票的差值
	pipeline := client.Pipeline()
	pipeline.ZIncrBy(getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID) // ZIncrBy为有序集合key中元素postID的分值加上op*diff*scorePerVote

	//3.记录用户为该帖子投过票的数据
	if value == 0 {
		pipeline.ZRem(getRedisKey(KeyPostVotedZSetPF+postID), postID) // ZRem在有序集合key中删除元素
	} else {
		pipeline.ZAdd(getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{ // ZAdd往有序集合key中加入带分值元素
			Score:  value, // 赞成or反对
			Member: userID,
		})
	}
	_, err := pipeline.Exec()
	return err
}
