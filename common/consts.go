package common

const (
	// chapter01
	// redis key

	// article id 集合
	ArticleID = "article:id"
	// 已投票集合前缀
	VotedSetPre = "voted:set:"
	// article 哈希集合前缀
	ArticleHashSetPre = "article:hash:set:"
	// article 得分有序集合
	Score = "score"
	// article 时间有序集合
	Time = "time"

	// 一周的秒数
	OneWeekInSeconds = 60 * 60 * 24 * 7
	// article 一票得分
	VotedScore = 432

	// chapter02

	// login 哈希集合
	LoginHash = "login"

	// 用户（token）最后访问时间的有序集合
	Recent = "recent"

	// 用户最近浏览商品的有序集合前缀
	ViewedPre = "viewed:"

	// 购物车前缀
	CartPre = "cart:"

	// 网页请求缓存前缀
	ReqCachePre = "reqcache:"

	// schedule 有序集合
	Schedule = "schedule"
	// delay 有序集合
	Delay = "delay"
)
