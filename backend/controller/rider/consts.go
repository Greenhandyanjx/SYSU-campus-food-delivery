package rider

const (
	OrderStatusPendingPay  = 1
	OrderStatusMerchantNew = 2 // ✅ 商家待接单（骑手不使用）
	OrderStatusToDeliver   = 3 // ✅ 待派送（商家已接单，骑手从这里开始）
	OrderStatusDelivering  = 4 // ✅ 派送中
	OrderStatusDone        = 5
	OrderStatusCanceled    = 6
)
