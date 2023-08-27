package api

import (
	"client/form"
	"client/global"
	"client/proto"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
	"time"
)

type CustomClaims struct {
	// 可根据需要自行添加字段
	Username string `json:"username"`

	jwt.RegisteredClaims // 内嵌标准的声明
}

const TokenExpireDuration = time.Hour * 24

var CustomSecret = []byte("夏天夏天悄悄过去")

// GenToken 生成JWT
func GenToken(username string) (string, error) {
	// 创建一个我们自己的声明
	claims := CustomClaims{
		username,
		// 自定义字段
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			Issuer:    "my-project", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(CustomSecret)
}
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析token
	// 如果是自定义Claim结构体则需要使用 ParseWithClaims 方法
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		// 直接使用标准的Claim则可以直接使用Parse方法
		//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return CustomSecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 对token对象中的Claim进行类型断言
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": e.Message(),
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.AlreadyExists:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": e.Message(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其他错误",
				})
			}
		}
	}
}
func UserRegister(c *gin.Context) {
	//conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserInfo.Host, global.ServerConfig.UserInfo.Port), grpc.WithInsecure())
	//if err != nil {
	//	zap.S().Errorw("连接服务端失败")
	//}
	//defer conn.Close()
	//userClient := proto.NewUserClient(conn)
	var user form.RegisterForm
	if err := c.ShouldBind(&user); err != nil {
		HandleGrpcErrorToHttp(err, c)
	}
	_, err := global.UserSrvClient.CreateUser(c, &proto.CreateUserRequest{
		Username: user.UserName,
		Password: user.PassWord,
		Price:    user.Price,
	})
	if err != nil {
		zap.S().Errorw("注册失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": "注册成功",
		})
	}

}
func UserLogin(c *gin.Context) {
	//conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.UserInfo.Host, global.ServerConfig.UserInfo.Port), grpc.WithInsecure())
	//if err != nil {
	//	zap.S().Errorw("连接服务端失败")
	//}
	//defer conn.Close()
	//userClient := proto.NewUserClient(conn)
	var user form.PassWordLoginForm
	if err := c.ShouldBind(&user); err != nil {
		HandleGrpcErrorToHttp(err, c)
	}
	resp, err := global.UserSrvClient.GetUserByName(c, &proto.NameRequest{Name: user.UserName})
	if err != nil {
		zap.S().Errorw("登陆失败")
		HandleGrpcErrorToHttp(err, c)
	}
	rsp, err := global.UserSrvClient.CheckPassword(c, &proto.PasswordCheckInfo{
		Password:          user.PassWord,
		EncryptedPassword: resp.Password,
	})
	if err != nil {
		zap.S().Errorw("登陆失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
	}
	if rsp.Success != true {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "密码错误",
		})
	} else {
		tokenstring, _ := GenToken(user.UserName)
		c.JSON(200, gin.H{
			"msg":   "成功登录",
			"token": tokenstring,
		})
	}

}

func SearchGoodsList(c *gin.Context) {
	//conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:51705"), grpc.WithInsecure())
	//if err != nil {
	//	zap.S().Errorw("连接服务端失败")
	//}
	//defer conn.Close()
	//userClient := proto.NewUserClient(conn)
	name := c.Query("title")
	resp, err := global.UserSrvClient.GetGoods(context.Background(), &proto.GetGoodsRequest{Title: name})
	if err != nil {
		zap.S().Errorw("搜索商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": resp,
		})
	}

}
func FavoriteGoods(c *gin.Context) {

	goods_id := c.Query("goods_id")
	goodsId, err := strconv.Atoi(goods_id)
	name, ok := c.Get("name")
	if !ok {
		c.JSON(200, gin.H{
			"msg": "没有登陆",
		})
	}
	resp, err := global.UserSrvClient.FavoriteGoods(context.Background(), &proto.GoodsUserRequest{
		Name:    name.(string),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		zap.S().Errorw("收藏商品失败")
		HandleGrpcErrorToHttp(err, c)
	}
	if resp.Success != true {
		zap.S().Errorw("收藏商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "成功收藏",
		})
	}

}
func GetFavoriteGoods(c *gin.Context) {

	name, _ := c.Get("name")
	resp, err := global.UserSrvClient.GetFavoriteGoods(context.Background(), &proto.NameRequest{Name: name.(string)})
	if err != nil {
		zap.S().Errorw("查询已收藏商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": resp,
		})
	}

}
func ClearGoodsInCar(c *gin.Context) {

	name, _ := c.Get("name")
	_, err := global.UserSrvClient.ClearGoodsInCar(context.Background(), &proto.NameRequest{Name: name.(string)})
	if err != nil {
		zap.S().Errorw("清空购物车失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": "清空购物车成功",
		})
	}

}
func BuyAllGoodsInCar(c *gin.Context) {

	name, _ := c.Get("name")
	_, err := global.UserSrvClient.BuyAllGoodsInCar(context.Background(), &proto.NameRequest{Name: name.(string)})
	if err != nil {
		zap.S().Errorw("购买全部购物车失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": "购买全部购物车商品成功",
		})
	}

}
func AddGoodsInCar(c *gin.Context) {

	name, _ := c.Get("name")
	goods_id := c.Query("goods_id")
	goodsId, err := strconv.Atoi(goods_id)
	resp, err := global.UserSrvClient.AddGoodsInCar(context.Background(), &proto.GoodsUserRequest{
		Name:    name.(string),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		zap.S().Errorw("在购物车添加商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg":      "在购物车添加商品成功",
			"quantity": resp.Quantity,
		})
	}

}
func DeleteGoodsInCar(c *gin.Context) {

	name, _ := c.Get("name")
	goods_id := c.Query("goods_id")
	goodsId, err := strconv.Atoi(goods_id)
	resp, err := global.UserSrvClient.ReduceGoodsInCar(context.Background(), &proto.GoodsUserRequest{
		Name:    name.(string),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		zap.S().Errorw("在购物车减少商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg":      "在购物车减少商品成功",
			"quantity": resp.Quantity,
		})
	}

}
func BuyGoods(c *gin.Context) {

	name, _ := c.Get("name")
	goods_id := c.Query("goods_id")
	goodsId, err := strconv.Atoi(goods_id)
	_, err = global.UserSrvClient.BuyGoods(context.Background(), &proto.GoodsUserRequest{
		Name:    name.(string),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		zap.S().Errorw("购买商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(200, gin.H{
			"msg": "购买商品成功",
		})
	}

}
func PublishGoods(c *gin.Context) {

	var goods form.GoodsForm
	err := c.ShouldBind(&goods)
	if err != nil {
		zap.S().Errorw("商品信息不完整")
		HandleGrpcErrorToHttp(err, c)
	}
	name, _ := c.Get("name")
	resp, err := global.UserSrvClient.PublishGoods(context.Background(), &proto.CreateGoodsRequest{
		Title:   goods.Title,
		Price:   goods.Price,
		Content: goods.Describe,
		Name:    name.(string),
	})
	if err != nil {
		zap.S().Errorw("发布商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"title":    resp.Title,
			"price":    resp.Price,
			"describe": resp.Content,
		})
	}

}
func DeleteGoods(c *gin.Context) {

	goods_id := c.Query("goods_id")
	goodsId, err := strconv.Atoi(goods_id)
	name, _ := c.Get("name")
	_, err = global.UserSrvClient.DeleteGoods(context.Background(), &proto.GoodsUserRequest{
		Name:    name.(string),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		zap.S().Errorw("删除商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg": "删除商品成功",
		})
	}

}
func UpdateGoods(c *gin.Context) {

	var goods form.UpdateGoodsForm
	err := c.ShouldBind(&goods)
	if err != nil {
		zap.S().Errorw("商品信息不完整")
		HandleGrpcErrorToHttp(err, c)
	}
	name, _ := c.Get("name")
	resp, err := global.UserSrvClient.UpdateGoods(context.Background(), &proto.UpdateGoodsRequest{
		Title:   goods.Title,
		Price:   goods.Price,
		Content: goods.Describe,
		Name:    name.(string),
		GoodsId: int32(goods.GoodsID),
	})
	if err != nil {
		zap.S().Errorw("更新商品失败")
		HandleGrpcErrorToHttp(err, c)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"title":    resp.Title,
			"price":    resp.Price,
			"describe": resp.Content,
		})
	}

}
