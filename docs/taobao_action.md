# Action

## Models


### `TbkScPublisherInfoGetReq` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`InfoType`|`int`|`info_type`|1	类型，必选 1:渠道信息；2:会员信息
`RelationId`|`int64`|`relation_id,omitempty`|2323	渠道独占 - 渠道关系ID
`PageNo`|`int`|`page_no,omitempty`|1	第几页
`PageSize`|`int`|`page_size,omitempty`|	10	每页大小
`SpecialId`|`string`|`special_id,omitempty`|1212	会员独占 - 会员运营ID
`ExternalId`|`string`|`external_id,omitempty`|	12345	淘宝客外部用户标记，如自身系统账户ID；微信ID等
`RelationApp`|`string`|`relation_app`|common	备案的场景：common（通用备案），etao（一淘备案），minietao（一淘小程序备案），offlineShop（线下门店备案），offlinePerson（线下个人备案）。如不填默认common。查询会员信息只需填写common即可


### `tbkScPublisherInfoGetResp` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkScPublisherInfoGetResponse`|`tbkScPublisherInfoGetResponse`|`tbk_sc_publisher_info_get_response`| 
|`*RespCommon`|| 


### `tbkScPublisherInfoGetResponse` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`Data`|`TbkScPublisherInfoGetResult`|`data`| 

### `TbkScPublisherInfoGetResult` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`RootPidChannelList`|`RootPidChannelList`|`root_pid_channel_list`|  ["mm_1_1_1"]	渠道专属pidList
`TotalCount     `|`   int           `|`   total_count`|40	共享字段 - 总记录数
`InviterList    `|`   InviterList   `|`   inviter_list`|inviterList	共享字段 - 渠道或会员列表

### `RootPidChannelList` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`String`|`[]string`|`string`| ["mm_1_1_1"]	渠道专属pidList

### `InviterList` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`MapData`|`[]MapData`|`map_data`| inviterList	共享字段 - 渠道或会员列表

### `MapData` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`MapData`|`[]MapData`|`map_data`| inviterList	共享字段 - 渠道或会员列表
`RelationApp  `|`string      `|`relation_app`|	common	共享字段 - 备案场景：common（通用备案），etao（一淘备案），minietao（一淘小程序备案），offlineShop（线下门店备案），offlinePerson（线下个人备案）
`CreateDate   `|`Time      `|`create_date`|2018-06-01 11:12:23	共享字段 - 备案日期
`AccountName  `|`string      `|`account_name`|s**x	共享字段 - 渠道/会员昵称
`RealName     `|`string      `|`real_name`|xxx	共享字段 - 渠道/会员姓名
`RelationID   `|`int         `|`relation_id`|40232	渠道独有 - 渠道关系ID
`OfflineScene `|`string      `|`offline_scene`|	门店	渠道独有 - 线下场景信息，1 - 门店，2- 学校，3 - 工厂，4 - 其他
`OnlineScene  `|`string      `|`online_scene`|微信群	渠道独有 - 线上场景信息，1 - 微信群，2- QQ群，3 - 其他
`Note         `|`string      `|`note`|小蜜蜂	渠道独有 - 媒体侧渠道备注信息
`RootPid      `|`string      `|`root_pid`|mm_1_1_1	共享字段 - 渠道/会员专属pid
`Rtag         `|`string      `|`rtag`|	123	共享字段 - 渠道/会员原始身份信息
`OfflineInfo  `|`OfflineInfo `|`offline_info`|	线下	线下备案专属信息
`SpecialID    `|`int         `|`special_id`|12345	会员独有 - 会员运营ID
`PunishStatus `|`string      `|`punish_status`|1	渠道独有 - 处罚状态
`ExternalID   `|`string      `|`external_id`|	12345	淘宝客外部用户标记，如自身系统账户ID；微信ID等

### `OfflineInfo` 淘宝客-公用-私域用户备案信息查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`MapData`|`[]MapData`|`map_data`| inviterList	共享字段 - 渠道或会员列表
`ShopName        `|`string `|`shop_name`|日用百货店	渠道独有 -店铺名称
`ShopType        `|`string `|`shop_type`|生活服务类 电信营业厅	渠道独有 -店铺类型
`PhoneNumber     `|`string `|`phone_number`|	1590000000	渠道独有 -手机号码
`DetailAddress   `|`string `|`detail_address`| xx街道xx号楼	渠道独有 -详细地址
`Location        `|`string `|`location`|	内蒙古自治区 呼和浩特市	渠道独有 -地区
`ShopCertifyType `|`string `|`shop_certify_type`|营业执照	渠道独有 -证件类型
`CertifyNumber   `|`string `|`certify_number`|23445677	渠道独有 -对应的证件证件类型编号
`Career          `|`string `|`career`|个人 快递员	渠道独有 -经营类型

### `TbkScPublisherInfoSaveReq` 淘宝客-公用-私域用户备案

Name|Type|JSON|Doc
:---|:---|:---|:--
`RelationFrom`|`string`|`relation_from,omitempty`| 1	渠道备案 - 来源，取链接的来源
`OfflineScene`|`string`|`offline_scene,omitempty`|	1	渠道备案 - 线下场景信息，1 - 门店，2- 学校，3 - 工厂，4 - 其他
`OnlineScene`|`string`|`online_scene,omitempty`|1	渠道备案 - 线上场景信息，1 - 微信群，2- QQ群，3 - 其他
`InviterCode`|`string`|`inviter_code`| xxx	渠道备案 - 淘宝客邀请渠道的邀请码
`InfoType`|`int`|`info_type`|	1	类型，必选 默认为1:
`Note`|`string`|`note,omitempty`|小蜜蜂	媒体侧渠道备注
`RegisterInfo`|`RawMessage`| `register_info,omitempty`|{"phoneNumber":"18801088599","city":"江苏省","province":"南京市","location":"玄武区花园小区","detailAddress":"5号楼3单元101室","shopType":"社区店","shopName":"全家便利店","shopCertifyType":"营业执照","certifyNumber":"111100299001"}	线下备案注册信息,字段包含: 电话号码(phoneNumber，必填),省(province,必填),市(city,必填),区县街道(location,必填),详细地址(detailAddress,必填),经营类型(career,线下个人必填),店铺类型(shopType,线下店铺必填),店铺名称(shopName,线下店铺必填),店铺证书类型(shopCertifyType,线下店铺选填),店铺证书编号(certifyNumber,线下店铺选填)

### `tbkScPublisherInfoSaveResp` 淘宝客-公用-私域用户备案

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkScPublisherInfoSaveResponse`|`tbkScPublisherInfoSaveResponse`|`tbk_sc_publisher_info_save_response`| 
|`*RespCommon`|| 


### `tbkScPublisherInfoSaveResponse` 淘宝客-公用-私域用户备案

Name|Type|JSON|Doc
:---|:---|:---|:--
`Data`|`TbkScPublisherInfoSaveResult`|`data`| 

### `TbkScPublisherInfoSaveResult` 淘宝客-公用-私域用户备案

Name|Type|JSON|Doc
:---|:---|:---|:--
`RelationId`|`int64`|`relation_id`|  40232	渠道关系ID
`AccountName`|`string`|`account_name`|  	xxx	渠道昵称
`SpecialId`|`int64`|`special_id`|  32304	会员运营ID
`Desc`|`int64`|`string`|  绑定成功	如果重复绑定会提示：”重复绑定渠道“或”重复绑定粉丝“


### `TbkScInvitecodeReq` 淘宝客邀请码生成-社交

Name|Type|JSON|Doc
:---|:---|:---|:--
`RelationID`|`int64`|`relation_id,omitempty`| 11	渠道关系ID
`RelationApp`|`string`|`relation_app`|common	渠道推广的物料类型
`CodeType	`|`int`|`code_type`|1	邀请码类型，1 - 渠道邀请，2 - 渠道裂变，3 -会员邀请


### `tbkScInvitecodeGetResp` 淘宝客邀请码生成-社交

Name|Type|JSON|Doc
:---|:---|:---|:--
`ScInvitecodeGetResponse`|`tbkScInvitecodeGetResponse`|`tbk_sc_invitecode_get_response`| 
|`*RespCommon`|| 


### `tbkScInvitecodeGetResponse` 淘宝客邀请码生成-社交

Name|Type|JSON|Doc
:---|:---|:---|:--
`Data`|`TbkScInvitecodeResult`|`data`| 

### `TbkScInvitecodeResult` 淘宝客邀请码生成-社交

Name|Type|JSON|Doc
:---|:---|:---|:--
`InviterCode `|`string`|`inviter_code`|  xxxx	邀请码

### `TbkItemInfoGetReq` 淘宝客商品详情查询（简版）

Name|Type|JSON|Doc
:---|:---|:---|:--
`NumIDs`|`string`|`num_iids`| 商品ID串，用,分割，最大40个
`Platform`|`int`|`platform,omitempty`|链接形式：1：PC，2：无线，默认：１
`Ip	`|`string`|`ip,omitempty`|ip地址，影响邮费获取，如果不传或者传入不准确，邮费无法精准提供


### `tbkitemInfoGetResp` 淘宝客商品详情查询（简版）

Name|Type|JSON|Doc
:---|:---|:---|:--
`ItemInfoGetResponse`|`tbkitemInfoGetResponse`|`tbk_item_info_get_response`| 
|`*RespCommon`|| 

### `tbkitemInfoGetResponse` 淘宝客商品详情查询（简版）

Name|Type|JSON|Doc
:---|:---|:---|:--
`Results`|`tbkitemInfoGetResults`|`results`| 
`RequestID`|`string`|`request_id`| 

### `tbkitemInfoGetResults` 淘宝客商品详情查询（简版）

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkItemInfoGetResults`|`[]TbkItemInfoGetResult`|`n_tbk_item`|  淘宝客商品


### `smallImages` 淘宝客商品详情查询（简版）

Name|Type|JSON|Doc
:---|:---|:---|:--
`String`|`[]string`|`string`|  淘宝客商品

### `TbkItemInfoGetResult` 淘宝客商品详情查询（简版）

Name|Type|Json|Doc
:---|:---|:---|:--
`CatName                    `|`string      `|`cat_name`|	女装	一级类目名称
`NumIid                     `|`int         `|`num_iid`|123	商品ID
`Title                      `|`string      `|`title`|连衣裙	商品标题
`PictURL                    `|`string      `|`pict_url`|http://gi4.md.alicdn.com/bao/uploaded/i4/xxx.jpg	商品主图
`SmallImages                `|`smallImages `|`small_images`|http://gi4.md.alicdn.com/bao/uploaded/i4/xxx.jpg	商品小图列表
`ReservePrice               `|`string      `|`reserve_price`|102.00	商品一口价格
`ZkFinalPrice               `|`string      `|`zk_final_price`|	88.00	折扣价（元） 若属于预售商品，付定金时间内，折扣价=预售价
`UserType                   `|`int         `|`user_type`|1	卖家类型，0表示集市，1表示商城
`Provcity                   `|`string      `|`provcity`|杭州	商品所在地
`ItemURL                    `|`string      `|`item_url`|http://detail.m.tmall.com/item.htm?id=xxx	商品链接
`SellerID                   `|`int         `|`seller_id`|123	卖家id
`Volume                     `|`int         `|`volume`|1	30天销量
`Nick                       `|`string      `|`nick`|xx旗舰店	店铺名称
`CatLeafName                `|`string      `|`cat_leaf_name`|情趣内衣	叶子类目名称
`IsPrepay                   `|`bool        `|`is_prepay`|true	是否加入消费者保障
`ShopDsr                    `|`int         `|`shop_dsr`|23	店铺dsr 评分
`Ratesum                    `|`int         `|`ratesum`|13	卖家等级
`IRfdRate                   `|`bool        `|`i_rfd_rate`|true	退款率是否低于行业均值
`HGoodRate                  `|`bool        `|`h_good_rate`|true	好评率是否高于行业均值
`HPayRate30                 `|`bool        `|`h_pay_rate30`|true	成交转化是否高于行业均值
`FreeShipment               `|`bool        `|`free_shipment`|true	是否包邮
`MaterialLibType            `|`string      `|`material_lib_type`|1	商品库类型，支持多库类型输出，以英文逗号分隔“,”分隔，1:营销商品主推库，2. 内容商品库，如果值为空则不属于1，2这两种商品类型
`PresaleDiscountFeeText     `|`string      `|`presale_discount_fee_text`|	付定金立减20元	预售商品-商品优惠信息
`PresaleTailEndTime         `|`int64       `|`presale_tail_end_time`|1937297392332	预售商品-付定金结束时间（毫秒）
`PresaleTailStartTime       `|`int64       `|`presale_tail_start_time`|1937297392332	预售商品-付尾款开始时间（毫秒）
`PresaleEndTime             `|`int64       `|`presale_end_time`|	1937297392332	预售商品-付定金结束时间（毫秒）
`PresaleStartTime           `|`int64       `|`presale_start_time`|1937297392332	预售商品-付定金开始时间（毫秒）
`PresaleDeposit             `|`string      `|`presale_deposit`|100	预售商品-定金（元）
`JuPlayEndTime              `|`int64       `|`ju_play_end_time`|1937297392332	聚划算满减 -结束时间（毫秒）    
`JuPlayStartTime            `|`int64       `|`ju_play_start_time`|	1937297392332	聚划算满减 -开始时间（毫秒）
`PlayInfo                   `|`string      `|`play_info`|玩法	1聚划算满减：满N件减X元，满N件X折，满N件X元） 2天猫限时抢：前N分钟每件X元，前N分钟满N件每件X元，前N件每件X元）
`TmallPlayActivityEndTime   `|`int64       `|`tmall_play_activity_end_time`|1937297392332	天猫限时抢可售 -结束时间（毫秒）
`TmallPlayActivityStartTime `|`int64       `|`tmall_play_activity_start_time`|1937297392332	天猫限时抢可售 -开始时间（毫秒）
`JuOnlineStartTime          `|`string      `|`ju_online_start_time`|	1581868800000	聚划算信息-聚淘开始时间（毫秒）
`JuOnlineEndTime            `|`string      `|`ju_online_end_time`|1582300799000	聚划算信息-聚淘结束时间（毫秒）
`JuPreShowStartTime         `|`string      `|`ju_pre_show_start_time`|1581868800000	聚划算信息-商品预热开始时间（毫秒）
`JuPreShowEndTime           `|`string      `|`ju_pre_show_end_time`|1582300799000	聚划算信息-商品预热结束时间（毫秒）
`SalePrice                  `|`string      `|`sale_price`|168	活动价
`KuadianPromotionInfo       `|`string      `|`kuadian_promotion_info`|["每100减20","每200减50"]	跨店满减信息

### `CouponGetReq` 阿里妈妈推广券信息查询。传入商品ID+券ID，或者传入me参数，均可查询券信息。

Name|Type|JSON|Doc
:---|:---|:---|:--
`Me`|`string`|`me,omitempty`| nfr%2BYTo2k1PX18gaNN%2BIPkIG2PadNYbBnwEsv6mRavWieOoOE3L9OdmbDSSyHbGxBAXjHpLKvZbL1320ML%2BCF5FRtW7N7yJ056Lgym4X01A%3D	带券ID与商品ID的加密串
`ItemID`|`int`|`item_id,omitempty`|123	商品ID
`ActivityID	`|`string`|`activity_id,omitempty`|sdfwe3eefsdf	券ID

### `couponGetResp` 淘宝客-公用-阿里妈妈推广券详情查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkCouponGetResponse`|`tbkCouponGetResponse`|`tbk_coupon_get_response`| 
|`*RespCommon`|| 

### `tbkCouponGetResponse` 淘宝客-公用-阿里妈妈推广券详情查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`Data`|`CouponGetResult`|`data`| 
`RequestID`|`string`|`request_id`| 

### `CouponGetResult` 淘宝客-公用-阿里妈妈推广券详情查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`CouponStartFee    `|`string `|`coupon_start_fee`|	29.00	优惠券门槛金额
`CouponRemainCount `|`int    `|`coupon_remain_count`|26996	优惠券剩余量
`CouponTotalCount  `|`int    `|`coupon_total_count`|	30000	优惠券总量
`CouponEndTime     `|`string `|`coupon_end_time`|2017-08-17	优惠券结束时间
`CouponStartTime   `|`string `|`coupon_start_time`|	2017-08-15	优惠券开始时间
`CouponAmount      `|`string `|`coupon_amount`|10.00	优惠券金额
`CouponSrcScene    `|`int    `|`coupon_src_scene`|	1	券类型，1 表示全网公开券，4 表示妈妈渠道券
`CouponType        `|`int    `|`coupon_type`|	0	券属性，0表示店铺券，1表示单品券
`CouponActivityID  `|`string `|`coupon_activity_id`|	xsdss	券ID

### `JuTqgGetReq` 获取淘抢购的数据，淘客商品转淘客链接，非淘客商品输出普通链接

Name|Type|JSON|Doc
:---|:---|:---|:--
`AdzoneID`|`int64`|`adzone_id`|123	推广位id（推广位申请方式：http://club.alimama.com/read.php?spm=0.0.0.0.npQdST&tid=6306396&ds=1&page=1&toread=1）
`Fields`|`string`|`fields`|click_url,pic_url,reserve_price,zk_final_price,total_amount,sold_num,title,category_name,start_time,end_time	需返回的字段列表
`StartTime`|`string`|`start_time`|2016-08-09 09:00:00	最早开团时间
`EndTime`|`string`|`end_time`|2016-08-09 16:00:00	最晚开团时间

`PageNO`|`int`|`page_no,omitempty`|1	第几页，默认1，1~100
`PageSize	`|`string`|`page_size,omitempty`|40	页大小，默认40，1~40

### `juTqgGetResp`  淘抢购api 

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkJuTqgGetResponse`|`tbkJuTqgGetResponse`|`tbk_ju_tqg_get_response`| 

|`*RespCommon`|| 

### `tbkJuTqgGetResponse`  淘抢购api

Name|Type|JSON|Doc
:---|:---|:---|:--
`Results`|`[]TbkJuTqgGetResult`|`results`| 
`TotalResults`|`int`|`total_results`| 	20	返回的结果数
`RequestID`|`string`|`request_id`| 

### `TbkJuTqgGetResult`  淘抢购api

Name|Type|JSON|Doc
:---|:---|:---|:--
`Title        `|`string `|`title`|连衣裙	商品标题
`TotalAmount  `|`int    `|`total_amount`|100	总库存
`ClickURL     `|`string `|`click_url`|http://s.click.taobao.com/t?e=x	商品链接（是淘客商品返回淘客链接，非淘客商品返回普通h5链接）
`CategoryName `|`string `|`category_name`|潮流女装	类目名称
`ZkFinalPrice `|`string `|`zk_final_price`|50.00	淘抢购活动价
`EndTime      `|`string `|`end_time`|	2016-08-09 13:00:00	结束时间
`SoldNum      `|`int    `|`sold_num`|50	已抢购数量
`StartTime    `|`string `|`start_time`|	2016-08-09 12:00:00	开团时间
`ReservePrice `|`string `|`reserve_price`|	100.00	商品原价
`PicURL       `|`string `|`pic_url`|http: //img4.tbcdn.cn/tfscom/i4/189490253156622336/TB2bZuSsVXXXXcNXXXXXXXXXXXX_!!0-juitemmedia.jpg	商品主图
`NumIid       `|`int    `|`num_iid`|123	商品ID

### `TbkTpwdCreateReq` 提供淘客生成淘口令接口，淘客提交口令内容、logo、url等参数，生成淘口令关键key如：￥SADadW￥，后续进行文案包装组装用于传播

Name|Type|JSON|Doc
:---|:---|:---|:--
`UserID`|`string`|`user_id,omitempty`|123	生成口令的淘宝用户ID
`Text`|`string`|`text`|长度大于5个字符	口令弹框内容
`URL`|`string`|`url`|https://uland.taobao.com/	口令跳转目标页
`Logo`|`string`|`logo,omitempty`|	https://uland.taobao.com/	口令弹框logoURL
`Ext	`|`string`|`ext,omitempty`|	{}	扩展字段JSON格式

### `tbkTpwdCreateResp`  淘宝客-公用-淘口令生成 

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkTpwdCreateResponse`|`tbkTpwdCreateResponse`|`tbk_tpwd_create_response`| 
|`*RespCommon`|| 

### `tbkTpwdCreateResponse`  淘宝客-公用-淘口令生成 

| Name        | Type                  | JSON         | Doc |
|:------------|:----------------------|:-------------|:----|
| `Data`      | `TbkTpwdCreateResult` | `data`       |     |
| `RequestID` | `string`              | `request_id` |     |

### `TbkTpwdCreateResult`  淘宝客-公用-淘口令生成 

Name|Type|JSON|Doc
:---|:---|:---|:--
`Model`|`string`|`model`| ￥AADPOKFz￥	password

### `ItemClickExtractReq` 从长链接或短链接中解析出open_iid

Name|Type|JSON|Doc
:---|:---|:---|:--
`ClickURL	`|`string`|`click_url	`| https://s.click.taobao.com/***	长链接或短链接

### `itemClickExtractResp` 淘宝客-公用-链接解析出商品id 

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkItemClickExtractResponse	`|`TbkItemClickExtractResult`|`tbk_item_click_extract_response	`| 
|`*RespCommon`|| 

### `TbkItemClickExtractResult` 淘宝客-公用-链接解析出商品id 

Name|Type|JSON|Doc
:---|:---|:---|:--
`ItemID	`|`string`|`item_id	`| 	123	商品id
`OpenIid	`|`string`|`open_iid	`| xxxxx	商品混淆id

### `TbkDgMaterialOptionalReq`  淘宝客-推广者-物料搜索 
| Name                | Type     | JSON                                   | Doc                                                                                                                                                                       |
|:--------------------|:---------|:---------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `StartDsr `         | `int`    | `start_dsr,omitempty`            | 10 商品筛选(特定媒体支持)-店铺dsr评分。筛选大于等于当前设置的店铺dsr评分的商品0-50000之间                                                                                 |
| `PageSize `         | `int`    | `page_size,omitempty`            | 20 页大小，默认20，1~100                                                                                                                                                  |
| `PageNo   `         | `int`    | `page_no,omitempty`              | 1 第几页，默认：１                                                                                                                                                        |
| `Platform  `        | `int`    | `platform,omitempty`             | 1  链接形式：1：PC，2：无线，默认：１                                                                                                                                     |
| `EndTkRate    `     | `int`    | `end_tk_rate,omitempty`          | 1234 商品筛选-淘客佣金比率上限。如：1234表示12.34%                                                                                                                        |
| `StartTkRate    `   | `int`    | `start_tk_rate,omitempty`        | 1234 商品筛选-淘客佣金比率下限。如：1234表示12.34%                                                                                                                        |
| `EndPrice`          | `int`    | `end_price,omitempty`            | 10   商品筛选-折扣价范围上限。单位：元                                                                                                                                    |
| `StartPrice`        | `int`    | `start_price,omitempty`          | 10   商品筛选-折扣价范围下限。单位：元                                                                                                                                    |
| `IsOverseas`        | `bool`   | `is_overseas,omitempty`          | false    商品筛选-是否海外商品。true表示属于海外商品，false或不设置表示不限                                                                                               |
| `IsTmall`           | `bool`   | `is_tmall,omitempty`             | false    商品筛选-是否天猫商品。true表示属于天猫商品，false或不设置表示不限                                                                                               |
| `Sort`              | `string` | `sort,omitempty`                 | tk_rate_des  排序_des（降序），排序_asc（升序），销量（total_sales），淘客佣金比率（tk_rate）， 累计推广量（tk_total_sales），总支出佣金（tk_total_commi），价格（price） |
| `Itemloc`           | `string` | `itemloc,omitempty`              | 杭州   商品筛选-所在地                                                                                                                                                    |
| `Cat`               | `string` | `cat,omitempty`                  | 16,18    商品筛选-后台类目ID。用,分割，最大10个，该ID可以通过taobao.itemcats.get接口获取到                                                                                |
| `Q`                 | `string` | `q,omitempty`                    | 女装   商品筛选-查询词                                                                                                                                                    |
| `MaterialId`        | `int`    | `material_id,omitempty`          | 2836 不传时默认物料id=2836；如果直接对消费者投放，可使用官方个性化算法优化的搜索物料id=17004                                                                              |
| `HasCoupon`         | `bool`   | `has_coupon,omitempty`           | false    优惠券筛选-是否有优惠券。true表示该商品有优惠券，false或不设置表示不限                                                                                           |
| `Ip`                | `string` | `ip,omitempty`                   | 13.2.33.4    ip参数影响邮费获取，如果不传或者传入不准确，邮费无法精准提供                                                                                                 |
| `AdzoneId`          | `int64`    | `adzone_id`                      | 12345678 mm_xxx_xxx_12345678三段式的最后一段数字                                                                                                                          |
| `NeedFreeShipment`  | `bool`   | `need_free_shipment,omitempty`   | true 商品筛选-是否包邮。true表示包邮，false或不设置表示不限                                                                                                               |
| `NeedPrepay`        | `bool`   | `need_prepay,omitempty`          | true 商品筛选-是否加入消费者保障。true表示加入，false或不设置表示不限                                                                                                     |
| `IncludePayRate30`  | `bool`   | `include_pay_rate_30,omitempty`  | true 商品筛选(特定媒体支持)-成交转化是否高于行业均值。True表示大于等于，false或不设置表示不限                                                                             |
| `IncludeGoodRate`   | `bool`   | `include_good_rate,omitempty`    | true 商品筛选-好评率是否高于行业均值。True表示大于等于，false或不设置表示不限                                                                                             |
| `IncludeRfdRate`    | `bool`   | `include_rfd_rate,omitempty`     | true 商品筛选(特定媒体支持)-退款率是否低于行业均值。True表示大于等于，false或不设置表示不限                                                                               |
| `NpxLevel`          | `int`    | `npx_level,omitempty`            | 2    商品筛选-牛皮癣程度。取值：1不限，2无，3轻微                                                                                                                         |
| `EndKaTkRate`       | `int`    | `end_ka_tk_rate,omitempty`       | 1234 商品筛选-KA媒体淘客佣金比率上限。如：1234表示12.34%                                                                                                                  |
| `StartKaTkRate`     | `int`    | `start_ka_tk_rate,omitempty`     | 1234 商品筛选-KA媒体淘客佣金比率下限。如：1234表示12.34%                                                                                                                  |
| `DeviceEncrypt`     | `string` | `device_encrypt,omitempty`       | MD5  智能匹配-设备号加密类型：MD5                                                                                                                                         |
| `DeviceValue`       | `string` | `device_value,omitempty`         | xxx  智能匹配-设备号加密后的值（MD5加密需32位小写）                                                                                                                       |
| `DeviceType`        | `string` | `device_type,omitempty`          | IMEI 智能匹配-设备号类型：IMEI，或者IDFA，或者UTDID（UTDID不支持MD5加密），或者OAID                                                                                       |
| `LockRateEndTime`   | `int64`  | `lock_rate_end_time,omitempty`   | 1567440000000    锁佣结束时间                                                                                                                                             |
| `LockRateStartTime` | `int64`  | `lock_rate_start_time,omitempty` | 1567440000000    锁佣开始时间                                                                                                                                             |
| `Longitude`         | `string` | `longitude,omitempty`            | 121.473701   本地化业务入参-LBS信息-经度                                                                                                                                  |
| `Latitude`          | `string` | `latitude,omitempty`             | 31.230370    本地化业务入参-LBS信息-纬度                                                                                                                                  |
| `CityCode`         | `string` | `city_code,omitempty`            | 310000   本地化业务入参-LBS信息-国标城市码，仅支持单个请求，请求饿了么卡券物料时，该字段必填。 （详细城市ID见：https://mo.m.taobao.com/page_2020010315120200508）         |
| `SellerIds`        | `string` | `seller_ids,omitempty`           | 1,2,3,4  商家id，仅支持饿了么卡券商家ID，支持批量请求1-100以内，多个商家ID使用英文逗号分隔                                                                                |


### `tbkDgMaterialOptionalResp`  淘宝客-推广者-物料搜索 

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkDgMaterialOptionalResponse	`|`tbkDgMaterialOptionalResponse`|`tbk_dg_material_optional_response`| 
|`*RespCommon`|| 

### `tbkDgMaterialOptionalResponse`  淘宝客-推广者-物料搜索 

Name|Type|JSON|Doc
:---|:---|:---|:--
`TotalResults`|`int`|`total_results`| 1212	搜索到符合条件的结果总数
`ResultList`|`tbkDgMaterialOptionalResults`|`result_list`| 
`RequestID`|`string`|`request_id`| 

### `tbkDgMaterialOptionalResults`  淘宝客-推广者-物料搜索 

Name|Type|JSON|Doc
:---|:---|:---|:--
`MapData`|`[]TbkDgMaterialOptionalResult`|`map_data`| 

### `TbkDgMaterialOptionalResult`  淘宝客-推广者-物料搜索 

Name|Type|JSON|Doc
:---|:---|:---|:--
`CouponStartTime        `|`string     `| `coupon_start_time`|2017-10-29	优惠券信息-优惠券开始时间
`CouponEndTime          `|`string     `| `coupon_end_time`|2017-10-29	优惠券信息-优惠券结束时间
`InfoDxjh               `|`RawMessage`| `info_dxjh`|{"19013551":"2850","74510538":"2550"}	商品信息-定向计划信息
`TkTotalSales           `|`string     `| `tk_total_sales`|11	商品信息-淘客30天推广量
`TkTotalCommi           `|`string     `| `tk_total_commi`|323	商品信息-月支出佣金(该字段废弃，请勿再用)
`CouponID               `|`string     `| `coupon_id`|d62db1ab8d9546b1bf0ff49bda5fc33b	优惠券信息-优惠券id
`NumIid                 `|`int64      `| `num_iid`|556633720749	商品信息-宝贝id(该字段废弃，请勿再用)
`Title                  `|`string     `| `title`|毛呢阔腿裤港味复古女裤子秋冬九分裤萝卜裤显瘦高腰韩版2017新款	商品信息-商品标题
`PictURL                `|`string     `| `pict_url`|https://img.alicdn.com/bao/uploaded/i4/745957850/TB1WzSRmV9gSKJjSspbXXbeNXXa_!!0-item_pic.jpg	商品信息-商品主图
`SmallImages            `|`smallImages`| `small_images`|https://img.alicdn.com/i4/3077291146/TB2NA3poDnI8KJjSszgXXc8ApXa_!!3077291146.jpg	商品信息-商品小图列表
`ReservePrice           `|`string     `| `reserve_price`|102.00	商品信息-商品一口价格
`ZkFinalPrice           `|`string     `| `zk_final_price`|88.00	折扣价（元） 若属于预售商品，付定金时间内，折扣价=预售价
`UserType               `|`int        `| `user_type`|1	店铺信息-卖家类型。0表示集市，1表示天猫
`Provcity               `|`string     `| `provcity`|杭州	商品信息-宝贝所在地
`ItemURL                `|`string     `| `item_url`|https://item.taobao.com/item.htm?id=564591813536	链接-宝贝地址
`IncludeMkt             `|`string     `| `include_mkt`|false	商品信息-是否包含营销计划
`IncludeDxjh            `|`string     `| `include_dxjh`|false	商品信息-是否包含定向计划
`CommissionRate         `|`string     `| `commission_rate`|1550表示15.5%	商品信息-佣金比率。1550表示15.5%
`Volume                 `|`int        `| `volume`|123	商品信息-30天销量（饿了么卡券信息-总销量）
`SellerID               `|`int        `| `seller_id`|232332	店铺信息-卖家id
`CouponTotalCount       `|`int        `| `coupon_total_count`|22323	优惠券信息-优惠券总量
`CouponRemainCount      `|`int        `| `coupon_remain_count`|111	优惠券信息-优惠券剩余量
`CouponInfo             `|`string     `| `coupon_info`|满299元减20元	优惠券信息-优惠券满减信息
`CommissionType         `|`string     `| `commission_type`|MKT表示营销计划，SP表示定向计划，COMMON表示通用计划	商品信息-佣金类型。MKT表示营销计划，SP表示定向计划，COMMON表示通用计划
`ShopTitle              `|`string     `| `shop_title`|xx旗舰店	店铺信息-店铺名称
`ShopDsr                `|`int        `| `shop_dsr`|13	店铺信息-店铺dsr评分
`CouponShareURL         `|`string     `| `coupon_share_url`|uland.xxx	链接-宝贝+券二合一页面链接
`URL                    `|`string     `| `url`|s.click.xxx	链接-宝贝推广链接
`LevelOneCategoryName   `|`string     `| `level_one_category_name`|女装	商品信息-一级类目名称
`LevelOneCategoryID     `|`int        `| `level_one_category_id`|20	商品信息-一级类目ID
`CategoryName           `|`string     `| `category_name`|连衣裙	商品信息-叶子类目名称
`CategoryID             `|`int        `| `category_id`|162201	商品信息-叶子类目id
`ShortTitle             `|`string     `| `short_title`|xxsd	商品信息-商品短标题
`WhiteImage             `|`string     `| `white_image`|https://img.alicdn.com/bao/uploaded/i4/745957850/TB1WzSRmV9gSKJjSspbXXbeNXXa_!!0-item_pic.jpg	商品信息-商品白底图
`Oetime                 `|`string     `| `oetime`|2018-08-21 11:23:43	拼团专用-拼团结束时间
`Ostime                 `|`string     `| `ostime`|2018-08-21 11:23:43	拼团专用-拼团开始时间
`JddNum                 `|`int        `| `jdd_num`|10	拼团专用-拼团几人团
`JddPrice               `|`string     `| `jdd_price`|5	拼团专用-拼团拼成价，单位元
`UvSumPreSale           `|`int        `| `uv_sum_pre_sale`|23	预售专用-预售数量
`XID                    `|`string     `| `x_id`|uESS0mv8JvfJRMKlIgCD5At9VuHVBXoqBRihfQlo4ybLiKygRlqiN4eoxoZDfe38uSogWy6GB971jD2N8tLuuc	链接-物料块id(测试中请勿使用)
`CouponStartFee         `|`string     `| `coupon_start_fee`|29.00	优惠券信息-优惠券起用门槛，满X元可用。如：满299元减20元
`CouponAmount           `|`string     `| `coupon_amount`|10.00	优惠券（元） 若属于预售商品，该优惠券付尾款可用，付定金不可用
`ItemDescription        `|`string     `| `item_description`|季凉被 全棉亲肤	商品信息-宝贝描述(推荐理由)
`Nick                   `|`string     `| `nick`|旗舰店	店铺信息-卖家昵称
`OrigPrice              `|`string     `| `orig_price`|25	拼团专用-拼团一人价（原价)，单位元
`TotalStock             `|`int        `| `total_stock`|5555	拼团专用-拼团库存数量
`SellNum                `|`int        `| `sell_num`|1111	拼团专用-拼团已售数量
`Stock                  `|`int        `| `stock`|4444	拼团专用-拼团剩余库存
`TmallPlayActivityInfo  `|`string     `| `tmall_play_activity_info`|前n件x折	营销-天猫营销玩法
`ItemID                 `|`int64      `| `item_id`|5678899993	商品信息-宝贝id
`RealPostFee            `|`string     `| `real_post_fee`|0.00	商品邮费
`LockRate               `|`string     `| `lock_rate`|1567440000000	锁住的佣金率
`LockRateEndTime        `|`int64      `| `lock_rate_end_time`|1567440000000	锁佣结束时间
`LockRateStartTime      `|`int64      `| `lock_rate_start_time`|1567440000000	锁佣开始时间
`PresaleDiscountFeeText `|`string     `| `presale_discount_fee_text`|付定金立减5元	预售商品-优惠
`PresaleTailEndTime     `|`int64      `| `presale_tail_end_time`|1567440000000	预售商品-付尾款结束时间（毫秒）
`PresaleTailStartTime   `|`int64      `| `presale_tail_start_time`|1567440000000	预售商品-付尾款开始时间（毫秒）
`PresaleEndTime         `|`int64      `| `presale_end_time`|	1567440000000	预售商品-付定金结束时间（毫秒）
`PresaleStartTime       `|`int64      `| `presale_start_time`|1567440000000	预售商品-付定金开始时间（毫秒）
`PresaleDeposit         `|`string     `| `presale_deposit`|100	预售商品-定金（元）
`YsylTljSendTime        `|`string     `| `ysyl_tlj_send_time`|2019-11-10 21:59:59	预售有礼-淘礼金发放时间
`YsylCommissionRate     `|`string     `| `ysyl_commission_rate`|2030（表示20.3%）	预售有礼-佣金比例（ 预售有礼活动享受的推广佣金比例，注：推广该活动有特殊分成规则，请详见：https://tbk.bbs.taobao.com/detail.html?appId=45301&postId=9334376 ）
`YsylTljFace            `|`string     `| `ysyl_tlj_face`|0.6	预售有礼-预估淘礼金（元）
`YsylClickURL           `|`string     `| `ysyl_click_url`|https://uland.taobao.com/coupon/edetail?e=nqUNB1NOF3Bt3vqbdXnGloankzPYmeEFkgNrw6YHQf9pZTj41Orn8MwBAs06HAOzqQomYNedOiHDYPmqkFXqLR0HgBdG%2FDDL%2F1M%2FBw7Sf%2FesGXLf%2BqX4cbeC%2F2cR0p0NlWH0%2BknxpnCJJP%2FQkZSsyo1HvKjXo4uz&pid=mm_26381042_12970066_52864659&af=1	预售有礼-推广链接
`YsylTljUseEndTime      `|`string     `| `ysyl_tlj_use_end_time`|2019-11-10 21:59:59	预售有礼-淘礼金使用结束时间
`YsylTljUseStartTime    `|`string     `| `ysyl_tlj_use_start_time`|2019-11-10 21:59:59	预售有礼-淘礼金使用开始时间
`SaleBeginTime          `|`string     `| `sale_begin_time`|1567440000000	本地化-销售开始时间
`SaleEndTime            `|`string     `| `sale_end_time`|1567440000000	本地化-销售结束时间
`Distance               `|`string     `| `distance`|300	本地化-到门店距离（米）
`UsableShopID           `|`string     `| `usable_shop_id`|10001	本地化-可用店铺id
`UsableShopName         `|`string     `| `usable_shop_name`|	饿了么卡券专营店	本地化-可用店铺名称
`SalePrice              `|`string     `| `sale_price`|168	活动价
`KuadianPromotionInfo   `|`string     `| `kuadian_promotion_info`|["每100减20","每200减50"]	跨店满减信息

### `TbkOrderDetailsGetReq` 淘宝客-推广者-所有订单查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`QueryType`|`int`|`query_type,omitempty`| 1 查询时间类型，1：按照订单淘客创建时间查询，2:按照订单淘客付款时间查询，3:按照订单淘客结算时间查询
`PositionIndex`|`string`|`position_index,omitempty`|	2222_334666	位点，除第一页之外，都需要传递；前端原样返回。
`PageSize	`|`int`|`page_size,omitempty`|20	页大小，默认20，1~100
`MemberType`|`int`|`member_type,omitempty`| 2	推广者角色类型,2:二方，3:三方，不传，表示所有角色
`TkStatus`|`int`|`tk_status,omitempty`|	12	淘客订单状态，12-付款，13-关闭，14-确认收货，3-结算成功;不传，表示所有状态
`EndTime	`|`Time`|`end_time`|2019-04-23 12:28:22	订单查询结束时间，订单开始时间至订单结束时间，中间时间段日常要求不超过3个小时，但如618、双11、年货节等大促期间预估时间段不可超过20分钟，超过会提示错误，调用时请务必注意时间段的选择，以保证亲能正常调用！
`StartTime`|`Time`|`start_time`| 2019-04-05 12:18:22	订单查询开始时间
`JumpType	`|`int`|`jump_type,omitempty`| 1	跳转类型，当向前或者向后翻页必须提供,-1: 向前翻页,1：向后翻页
`PageNo`|`int`|`page_no,omitempty`| 	1	第几页，默认1，1~100
`OrderScene`|`int`|`order_scene,omitempty`| 	1	场景订单场景类型，1:常规订单，2:渠道订单，3:会员运营订单，默认为1


### `tbkOrderDetailsGetResp` 淘宝客-推广者-所有订单查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkOrderDetailsGetResponse`|`tbkOrderDetailsGetResponse`|`tbk_order_details_get_response`| 
|`*RespCommon`|| 

### `tbkOrderDetailsGetResponse` 淘宝客-推广者-所有订单查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`Data`|`tbkOrderDetailsGetData`|`data`| 
`RequestID`|`string`|`request_id`| 

### `tbkOrderDetailsGetData` 淘宝客-推广者-所有订单查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`Results      `|` tbkOrderDetailsGetResults `|`results`|
`HasPre       `|` bool    `|`has_pre`| false	是否还有上一页
`PositionIndex`|` string  `|`position_index`|1555904214_lGltNdNvSX2|1555917305_lJfPMeFmdt2	位点字段，由调用方原样传递
`HasNext      `|` bool    `|`has_next`|
`PageNo       `|` int     `|`page_no`| 1	页码
`PageSize     `|` int     `|`page_size`|11	页大小

### `tbkOrderDetailsGetResults` 淘宝客-推广者-所有订单查询

Name|Type|JSON|Doc
:---|:---|:---|:--
`TbkOrderDetailsGetResults`|`[]TbkOrderDetailsGetResult`|`publisher_order_dto`|  

### `ServiceFeeDtoList` 淘宝客-推广者-所有订单查询
Name|Type|JSON|Doc
:---|:---|:---|:--
`ServiceFeeDto`|`[]ServiceFeeDto`|`service_fee_dto`|  

### `ServiceFeeDto` 淘宝客-推广者-所有订单查询
Name|Type|JSON|Doc
:---|:---|:---|:--
`ShareRelativeRate `|`string `|`share_relative_rate`|	0.10	专项服务费率
`ShareFee          `|`string `|`share_fee`          | 	11.11	结算专项服务费
`SharePreFee       `|`string `|`share_pre_fee`      | 11.11	预估专项服务费
`TkShareRoleType   `|`int    `|`tk_share_role_type` | 	122	专项服务费来源，122-渠道

### `TbkOrderDetailsGetResult` 淘宝客-推广者-所有订单查询
Name|Type|JSON|Doc
:---|:---|:---|:--
`TbPaidTime                        `|` Time            `|`tb_paid_time`                             |2019-04-22 15:15:05	订单在淘宝拍下付款的时间
`TkPaidTime                        `|` Time            `|`tk_paid_time`                             | 2019-04-22 15:15:05	订单付款的时间，该时间同步淘宝，可能会略晚于买家在淘宝的订单创建时间
`PayPrice                          `|` string            `|`pay_price`                                | 9.11	买家确认收货的付款金额（不包含运费金额）
`PubShareFee                       `|` string            `|`pub_share_fee`                            | 0 结算预估收入=结算金额*提成。以买家确认收货的付款金额为基数，预估您可能获得的收入。因买家退款、您违规推广等原因，可能与您最终收入不一致。最终收入以月结后您实际收到的为准
`TradeID                           `|` string            `|`trade_id`                                 | 294159887445064307	买家通过购物车购买的每个商品对应的订单编号，此订单编号并未在淘宝买家后台透出
`TkOrderRole                       `|` int               `|`tk_order_role`                            | 	2	二方：佣金收益的第一归属者； 三方：从其他淘宝客佣金中进行分成的推广者
`TkEarningTime                     `|` Time            `|`tk_earning_time`                          | 2019-04-22 15:15:05	订单确认收货后且商家完成佣金支付的时间
`AdzoneID                          `|` int64             `|`adzone_id`                                | 11	推广位管理下的推广位名称对应的ID，同时也是pid=mm_1_2_3中的“3”这段数字
`PubShareRate                      `|` string            `|`pub_share_rate`                           | 100	从结算佣金中分得的收益比率
`Unid                              `|` string            `|`unid`                                     | 11	unid
`RefundTag                         `|` int               `|`refund_tag`                               | 	0	维权标签，0 含义为非维权 1 含义为维权订单
`SubsidyRate                       `|` string            `|`subsidy_rate`                             | 0	平台给与的补贴比率，如天猫、淘宝、聚划算等
`TkTotalRate                       `|` string            `|`tk_total_rate`                            | 	9.99	提成=收入比率*分成比率。指实际获得收益的比率
`ItemCategoryName                  `|` string            `|`item_category_name`                       | 淘小铺	商品所属的根类目，即一级类目的名称
`SellerNick                        `|` string            `|`seller_nick`                              | --	掌柜旺旺
`PubID                             `|` int64             `|`pub_id`                                   | 	98836808	推广者的会员id
`AlimamaRate                       `|` string            `|`alimama_rate`                             | 0	推广者赚取佣金后支付给阿里妈妈的技术服务费用的比率
`SubsidyType                       `|` string            `|`subsidy_type`                             | 	平台出资方，如天猫、淘宝、或聚划算等
`ItemImg                           `|` string            `|`item_img`                                 |img.alicdn.com/bao/uploaded/i1/2200782262419/O1CN01b5qlop1TjwarUo8fo_!!2200782262419.jpg	商品图片
`PubSharePreFee                    `|` string            `|`pub_share_pre_fee`                        | 0	付款预估收入=付款金额*提成。指买家付款金额为基数，预估您可能获得的收入。因买家退款等原因，可能与结算预估收入不一致
`AlipayTotalPrice                  `|` string            `|`alipay_total_price`                       | 11.22	买家拍下付款的金额（不包含运费金额）
`ItemTitle                         `|` string            `|`item_title`                               | tsh_rj_测试请不要拍_阶佣11.1	商品标题
`SiteName                          `|` string            `|`site_name`                                | 	合伙人	媒体管理下的对应ID的自定义名称
`ItemNum                           `|` int               `|`item_num`                                 | 2	商品数量
`SubsidyFee                        `|` string            `|`subsidy_fee`                              |	0	补贴金额=结算金额*补贴比率
`AlimamaShareFee                   `|` string            `|`alimama_share_fee`                        | 0	技术服务费=结算金额*收入比率*技术服务费率。推广者赚取佣金后支付给阿里妈妈的技术服务费用
`TradeParentID                     `|` string            `|`trade_parent_id`                          | 294159887445064307	买家在淘宝后台显示的订单编号
`OrderType                         `|` string            `|`order_type`                               | 如意淘	订单所属平台类型，包括天猫、淘宝、聚划算等
`TkCreateTime                      `|` Time            `|`tk_create_time`                           | 2019-04-22 15:15:05	订单创建的时间，该时间同步淘宝，可能会略晚于买家在淘宝的订单创建时间
`FlowSource                        `|` string            `|`flow_source`                              | 	--	产品类型
`TerminalType                      `|` string            `|`terminal_type`                            | 无线	成交平台
`ClickTime                         `|` Time            `|`click_time`                               | 2019-04-22 15:14:55	通过推广链接达到商品、店铺详情页的点击时间
`TkStatus                          `|` int               `|`tk_status`                                | 13	已付款：指订单已付款，但还未确认收货 已收货：指订单已确认收货，但商家佣金未支付 已结算：指订单已确认收货，且商家佣金已支付成功 已失效：指订单关闭/订单佣金小于0.01元，订单关闭主有：1）买家超时未付款； 2）买家付款前，买家/卖家取消了订付款后发起售中退款成功；3：订单结算，12：订单付款， 13订单失效，14：订单成功
`ItemPrice                         `|` string            `|`item_price`                               | 2.1	商品单价
`ItemID                            `|` int64             `|`item_id`                                  | 590141576510	商品id
`AdzoneName                        `|` string            `|`adzone_name`                              | 	推广位管理下的自定义推广位名称
`TotalCommissionRate               `|` string            `|`total_commission_rate`                    | 9.99	佣金比率
`ItemLink                          `|` string            `|`item_link`                                | 		商品链接
`SiteID                            `|` int               `|`site_id`                                  | 45598009	媒体管理下的ID，同时也是pid=mm_1_2_3中的“2”这段数字
`SellerShopTitle                   `|` string            `|`seller_shop_title`                        | --	店铺名称
`IncomeRate                        `|` string            `|`income_rate`                              | 9.99	订单结算的佣金比率+平台的补贴比率
`TotalCommissionFee                `|` string            `|`total_commission_fee`                     | 0	佣金金额=结算金额*佣金比率
`TkCommissionPreFeeForMediaPlatform`|` string            `|`tk_commission_pre_fee_for_media_platform` | 1.05	预估内容专项服务费：内容场景专项技术服务费，内容推广者在内容场景进行推广需要支付给阿里妈妈专项的技术服务费用。专项服务费＝付款金额＊专项服务费率。
`TkCommissionFeeForMediaPlatform   `|` string            `|`tk_commission_fee_for_media_platform`     | 1.05	结算内容专项服务费：内容场景专项技术服务费，内容推广者在内容场景进行推广需要支付给阿里妈妈专项的技术服务费用。专项服务费＝结算金额＊专项服务费率。
`TkCommissionRateForMediaPlatform  `|` string            `|`tk_commission_rate_for_media_platform`    | 0.01	内容专项服务费率：内容场景专项技术服务费率，内容推广者在内容场景进行推广需要按结算金额支付一定比例给阿里妈妈作为内容场景专项技术服务费，用于提供与内容平台实现产品技术对接等服务。
`SpecialID                         `|` int64             `|`special_id`                               | 	21321	会员运营id
`RelationID                        `|` int64             `|`relation_id`                              | 123123	渠道关系id
`DepositPrice                     `|` string            `|`deposit_price`                          | 22.32	预售金额，用户对预售商品支付的定金金额
`TbDepositTime                     `|` Time            `|`tb_deposit_time`                          | 	2019-09-09 12:01:01	预售时期，用户对预售商品支付定金的付款时间
`AlscID                            `|` string            `|`alsc_id`             | 32434	口碑子订单号
`AlscPid                           `|` string            `|`alsc_pid`            |  	3434	口碑父订单号
`ServiceFeeDtoList                 `|` ServiceFeeDtoList `|`service_fee_dto_list`| 	服务费信息