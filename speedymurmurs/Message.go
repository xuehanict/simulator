package speedymurmurs



/*************消息的多个类型*************/
/**
交易请求信息
*/
type payReq struct {
	requestID RequestID
	root      RouteID
	sender    RouteID
	dest      string
	value     float64
	upperHop  RouteID
}

type payRes struct {
	requestID RequestID
	root      RouteID
	sender    RouteID
	success   bool
	val       float64
}

/**
进行支付时,传递此消息
*/
type Payment struct {
	requestID RequestID
	root      RouteID
}

/**
和邻居进行地址交换时的请求消息类型
*/
type addrReq struct {
	reqSrc  RouteID
	reqRoot RouteID
	reqID   RequestID
}

/**
和邻居进行地址交换时的回复消息类型
*/
type addrRes struct {
	resSrc  RouteID
	resRoot RouteID
	reqID   RequestID
	addr    string
}

/**
通知邻居reset地址的消息类型
 */
type addrResetNoti struct {
	root RouteID
	src RouteID
}

/*************************************/


