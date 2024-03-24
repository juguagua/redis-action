package common

const (
	// chapter01
	// redis key

	// article id 集合
	ArticleID = "article-id"
	// 已投票集合前缀
	VotedSetPre = "voted-set:"
	// article 哈希集合前缀
	ArticleHashSetPre = "article-hash-set:"
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

	// 某个用户最近浏览商品的有序集合前缀
	ViewedPre = "viewed:"

	// 所有用户浏览的商品次数集合
	Viewed = "viewed"

	// 购物车前缀
	CartPre = "cart:"

	// 网页请求缓存前缀
	ReqCachePre = "reqcache:"

	// schedule 有序集合
	Schedule = "schedule"
	// delay 哈希集合
	Delay = "delay"
	// 数据行缓存前缀
	RowCachePre = "row-req:"

	// 买卖市场商品的有序集合
	Market = "Market"
	// 用户散列前缀
	UserPre = "user:"
	// 用户背包集合前缀
	UserBagPre = "user-bag:"

	// 最新日志列表前缀
	RecentLogListPre = "recent-log-list:"
	// 日志出现频率有序集合前缀
	FrequencyLogZSetPre = "frequency-log-zset:"
)
